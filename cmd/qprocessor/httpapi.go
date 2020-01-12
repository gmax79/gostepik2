package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"project/api"
	"project/systemstop"
)

type httpAPI struct {
	//mux             *http.ServeMux
	host    string
	graders []string
	tasks   controller
}

func (h *httpAPI) processed(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()
	id := keys.Get("id")
	fmt.Println("Processed task:", id)
	h.tasks.deleteTask(id)
	api.Response(w, "")
}

func (h *httpAPI) Serve(host string, graders []string) error {
	stop := systemstop.Subscribe()
	defer stop.Done()
	h.host = host
	h.graders = graders
	h.tasks = controller{}
	mux := http.NewServeMux()
	mux.HandleFunc("/processed", h.processed)
	server := http.Server{Addr: host, Handler: mux}
	go func() {
		<-stop.Signal()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()
	go func() {

	}()
	err := server.ListenAndServe()
	return err
}

func (h *httpAPI) SendTaskToGrader(mq *api.MQmessage) error {
	if mq.Source == "" || mq.Result == "" || mq.Testapp == "" {
		return fmt.Errorf("Missing one or more parameters in rmq message")
	}
	// find free grader
	for _, graderurl := range h.graders {
		free, err := h.getGraderStatus(graderurl)
		if err != nil {
			fmt.Printf("Grader %s in error state: %v\n", graderurl, err)
			continue
		}
		if free {
			id := genUUID()
			taskgm := &api.GraderMessage{MQmessage: *mq, Processed: "http://" + h.host + "/processed?id=" + id}
			url := "http://" + graderurl + "/newtask"
			err := api.MarshaledPostRequest(url, taskgm)
			if err != nil {
				fmt.Printf("Grader error in send task: %v\n", err)
				continue
			}
			h.tasks.addTask(id, taskgm)
			return nil
		}
	}
	return fmt.Errorf("Task not sent, no free grader")
}

func (h *httpAPI) getGraderStatus(graderaddr string) (bool, error) {
	url := "http://" + graderaddr + "/status"
	r, err := http.Get(url)
	if err != nil {
		return false, err
	}
	data, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return false, err
	}
	status := &api.GraderStatus{}
	err = json.Unmarshal(data, status)
	if err != nil {
		return false, err
	}
	if status.Status == "free" {
		return true, nil
	}
	if status.Status == "busy" {
		return false, nil
	}
	return false, fmt.Errorf("Uknown grader state: %s", status.Status)
}
