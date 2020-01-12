package httptemplates

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
)

// Factory - templates factory
type Factory struct {
	tmpl    *template.Template
	funcMap template.FuncMap
}

func Initialize() *Factory {
	t := template.New("")
	funcs := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"inc": func(i int) int {
			return i + 1
		},
	}
	return &Factory{tmpl: t, funcMap: funcs}
}

func (f *Factory) Execute(w http.ResponseWriter, name string, data interface{}) error {
	t := f.tmpl.Lookup(name)
	//if t == nil {
	file, err := os.Open("web/" + name + ".html")
	if err != nil {
		return err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	t, err = template.New(name).Funcs(f.funcMap).Parse(string(b))
	if err != nil {
		return fmt.Errorf("Template '%s' not exists", name)
	}
	//}
	var buffer bytes.Buffer
	if err := t.Execute(&buffer, data); err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Write(buffer.Bytes())
	return nil
}

func (f *Factory) Error(w http.ResponseWriter, statusCode int) {
	text := fmt.Sprintf("Error: %d", statusCode)
	w.WriteHeader(statusCode)
	w.Write([]byte(text))
}
