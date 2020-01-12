package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type MQmessage struct {
	Source  string `json:"source"`
	Result  string `json:"result"`
	Testapp string `json:"testapp"`
}

type GraderMessage struct {
	MQmessage
	Processed string `json:"processed"`
}

type GraderStatus struct {
	Status string `json:"status"`
}

type GraderResult struct {
	Result string `json:"result"`
}

func Response(w http.ResponseWriter, v interface{}) {
	switch v.(type) {
	case nil:
		w.WriteHeader(http.StatusNotFound)
	case error:
		err := v.(error)
		w.WriteHeader(http.StatusInternalServerError)
		errinfo := []byte(err.Error())
		w.Write(errinfo)
		return
	case string:
		str := v.(string)
		w.WriteHeader(http.StatusOK)
		if str != "" {
			w.Write([]byte(str))
		}
	case []byte:
		b := v.([]byte)
		w.WriteHeader(http.StatusOK)
		if len(b) > 0 {
			w.Write(b)
		}
	default:
		answer, err := json.Marshal(v)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errinfo := []byte(err.Error())
			w.Write(errinfo)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(answer)
		}
	}
}

func UnmarshalPostRequest(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return false
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Response(w, err)
		return false
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		Response(w, err)
		return false
	}
	return true
}

func MarshaledPostRequest(url string, v interface{}) error {
	fmt.Println("POST", url)
	attepts := 3
	jsonbytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(jsonbytes)
	apptype := "application/json"
	request, err := http.Post(url, apptype, buffer)
	for err != nil {
		attepts--
		if attepts == 0 {
			break
		}
		time.Sleep(time.Second * 5)
		request, err = http.Post(url, apptype, buffer)
	}
	if err == nil && request.StatusCode != http.StatusOK {
		return fmt.Errorf("Remote host returned status code: %d", request.StatusCode)
	}
	return err
}

func GetRequest(url string) error {
	fmt.Println("GET", url)
	attepts := 3
	_, err := http.Get(url)
	for err != nil {
		attepts--
		if attepts == 0 {
			break
		}
		time.Sleep(time.Second * 5)
		_, err = http.Get(url)
	}
	return err
}
