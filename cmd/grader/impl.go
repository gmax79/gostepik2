package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"project/api"
	"sync"
	"syscall"
	"time"
)

type grader struct {
	mux *http.ServeMux
	//message  api.GraderMessage
	mutex    *sync.Mutex
	free     bool
	finished chan struct{}
}

func createGrader() *grader {
	g := &grader{}
	mux := http.NewServeMux()
	mux.HandleFunc("/newtask", g.newtask)
	mux.HandleFunc("/status", g.status)
	g.mux = mux
	g.mutex = &sync.Mutex{}
	g.finished = make(chan struct{})
	return g
}

func (g *grader) Serve(host string) error {
	g.free = true
	go func() {
		for {
			select {
			case <-g.finished:
				g.mutex.Lock()
				g.free = true
				g.mutex.Unlock()
			}
		}
	}()
	s := http.Server{Addr: host, Handler: g.mux}
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		s.Shutdown(ctx)
	}()
	return s.ListenAndServe()
}

// health check and status
func (g *grader) status(w http.ResponseWriter, r *http.Request) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	s := api.GraderStatus{}
	if !g.free {
		s.Status = "busy"
	} else {
		s.Status = "free"
	}
	api.Response(w, s)
}

// accept new task from queue processor
func (g *grader) newtask(w http.ResponseWriter, r *http.Request) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if !g.free {
		api.Response(w, api.GraderStatus{Status: "busy"})
		return
	}
	msg := api.GraderMessage{}
	if !api.UnmarshalPostRequest(w, r, &msg) {
		return
	}
	g.free = false
	go processMessage(msg, g.finished)
	api.Response(w, api.GraderStatus{Status: "ok"})
}

func processMessage(msg api.GraderMessage, finished chan<- struct{}) {
	fmt.Println("[Grader] Got task:", msg)
	defer func() {
		finished <- struct{}{}
	}()

	//todo emulate working
	time.Sleep(time.Second * 5)
	// emulate check result
	rnd := rand.Intn(100)
	result := "ok"
	if (rnd % 2) == 0 {
		result = "error, try again"
	}
	fmt.Println("[Grader] Task result:", result)

	// send result back in qprocess
	err := api.GetRequest(msg.Processed)
	for err != nil {
		fmt.Println("[Grader]", err.Error())
		time.Sleep(time.Minute)
		err = api.GetRequest(msg.Processed)
	}

	// send result back to server
	err = api.MarshaledPostRequest(msg.Result, result)
	for err != nil {
		fmt.Println("[Grader]", err.Error())
		time.Sleep(time.Minute)
		err = api.MarshaledPostRequest(msg.Result, result)
	}
	fmt.Println("[Grader] Task processed")
}
