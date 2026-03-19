package api

import (
	"html/template"
	"log"
	"net/http"
)

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	log.Println("Serving index page")
	templ, err := template.ParseFiles("frontend/index.html")
	if err != nil {
		log.Println("index not found")
		return
	}
	err = templ.Execute(w, nil)
	if err != nil {
		log.Println("index not serve")
	}
}

func ServeExecution(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	templ, err := template.ParseFiles("frontend/execution.html")
	if err != nil {
		log.Println("execution page not found")
		return
	}
	err = templ.Execute(w, nil)
	if err != nil {
		log.Println("execution page not serve")
	}
}

func ServeExecute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	templ, err := template.ParseFiles("frontend/execute.html")
	if err != nil {
		log.Println("execute page not found")
		return
	}
	err = templ.Execute(w, nil)
	if err != nil {
		log.Println("execute page not serve")
	}
}
func ServeWorkflowDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	id := r.URL.Query().Get("id") // get ?id=xxxx
	templ, err := template.ParseFiles("frontend/workflow-detail.html")
	if err != nil {
		log.Println(err)
		return
	}
	data := struct {
		WorkflowID string
	}{
		WorkflowID: id,
	}
	templ.Execute(w, data)
}
