package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"project/api"
	"project/server/httptemplates"
	"project/server/provider"

	"github.com/streadway/amqp"
)

type mainHandler struct {
	host    string
	tmpl    *httptemplates.Factory
	content provider.CourseHandler
}

func getLessonID(r *http.Request) provider.LessonIndex {
	id := provider.LessonIndex{}
	id.CourseID = r.URL.Query().Get("id")
	id.LessonID = r.URL.Query().Get("l")
	return id
}

func getLessonUserLogin(r *http.Request) string {
	return "petrov"
}

func getLessonTestApp(r *http.Request) string {
	return "Go"
}

func (h *mainHandler) clist(w http.ResponseWriter, r *http.Request) {
	courses := h.content.GetCources()
	err := h.tmpl.Execute(w, "clist", courses)
	if err != nil {
		fmt.Println(err)
	}
}

func (h *mainHandler) cdetails(w http.ResponseWriter, r *http.Request) {
	id := getLessonID(r)
	lessons, ok := h.content.GetLessons(id.CourseID)
	if !ok {
		h.tmpl.Error(w, http.StatusNotFound)
		return
	}
	err := h.tmpl.Execute(w, "cdetails", lessons)
	if err != nil {
		fmt.Println(err)
	}
}

func (h *mainHandler) lesson(w http.ResponseWriter, r *http.Request) {
	id := getLessonID(r)
	lesson, ok := h.content.GetLesson(id)
	if !ok {
		h.tmpl.Error(w, http.StatusNotFound)
		return
	}
	err := h.tmpl.Execute(w, "lesson", lesson)
	if err != nil {
		fmt.Println(err)
	}
}

func (h *mainHandler) uploadReport(w http.ResponseWriter, r *http.Request) {
	id := getLessonID(r)
	user := getLessonUserLogin(r)
	testapp := getLessonTestApp(r)

	r.ParseMultipartForm(100 * 1024)
	file, handler, err := r.FormFile("report")
	if err != nil {
		h.tmpl.Error(w, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// save attachment into temp file first
	const timeformat = "02_Jan_06_15:04"
	currenttime := time.Now().Format(timeformat)
	idparts := []string{user, id.CourseID, id.LessonID, currenttime, handler.Filename}
	idfile := strings.Join(idparts, "-")
	filepath := "/tmp/" + idfile + ".bin"
	f, err := os.Create(filepath)
	writer := bufio.NewWriter(f)
	_, err = io.Copy(writer, file)
	f.Sync()
	f.Close()
	if err != nil {
		h.tmpl.Error(w, http.StatusInternalServerError)
		return
	}

	// registry file in provider (sql)
	report := provider.LessonReport{LessonIndex: id, Login: user, TmpFile: filepath}
	h.content.SaveReport(report)

	// send message into mq
	m := api.MQmessage{
		Source:  h.host + "/getreport?id=" + idfile,
		Result:  h.host + "/result?id=" + idfile,
		Testapp: testapp,
	}
	binm, err := json.Marshal(m)
	if err != nil {
		h.tmpl.Error(w, http.StatusInternalServerError)
		return
	}

	sendMQ(binm)

	err = h.tmpl.Execute(w, "uploaded", nil)
	if err != nil {
		h.tmpl.Error(w, http.StatusInternalServerError)
		return
	}
}

func (h *mainHandler) taskResult(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[Server] Task result ", r.URL.Host)
}

func sendMQ(data []byte) {
	var rabbitConn *amqp.Connection
	var rabbitChan *amqp.Channel
	var err error

	rabbitAddr := "amqp://grader:Vfnhtirf79@do.cc3.ru:5672/"
	rabbitConn, err = amqp.Dial(rabbitAddr)
	if err != nil {
		//panicOnError("cant connect to rabbit", err)
		return
	}
	rabbitChan, err = rabbitConn.Channel()
	if err != nil {
		//panicOnError("cant open chan", err)
		return
	}
	defer rabbitChan.Close()

	//data, _ := json.Marshal(ImgResizeTask{handler.Filename, md5Sum})

	fmt.Println("[Server] Put task ", string(data))

	err = rabbitChan.Publish(
		"",       // exchange
		"grader", // routing key
		false,    // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         data,
		})

}

func main() {
	var err error
	defer func() {
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}()

	var settingsJSON []byte
	if settingsJSON, err = ioutil.ReadFile("settings.json"); err != nil {
		return
	}
	var contentProvider provider.CourseHandler
	if contentProvider, err = provider.Create(settingsJSON); err != nil {
		return
	}

	handler := &mainHandler{host: "http://127.0.0.1:9090", tmpl: httptemplates.Initialize(), content: contentProvider}
	go func() {
		apiServer := http.NewServeMux()
		apiServer.HandleFunc("/result", handler.taskResult)
		fmt.Println("starting api server at :9090")
		err := http.ListenAndServe(":9090", apiServer)
		if err != nil {
			fmt.Println(err)
		}
	}()
	mainServer := http.NewServeMux()
	mainServer.HandleFunc("/", handler.clist)
	mainServer.HandleFunc("/cdetails", handler.cdetails)
	mainServer.HandleFunc("/lesson", handler.lesson)
	mainServer.HandleFunc("/uploadreport", handler.uploadReport)
	fmt.Println("starting main server at :8080")
	err = http.ListenAndServe(":8080", mainServer)
}
