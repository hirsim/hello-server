package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	rootPath             = "/"
	helloStringPath      = "/hello"
	helloHTMLPath        = "/hello.html"
	helloJSONPath        = "/hello.json"
	slothHelloStringPath = "/sloth/hello"
	slothHelloHTMLPath   = "/sloth/hello.html"
	slothHelloJSONPath   = "/sloth/hello.json"
)

var (
	printText = "Hello World!"

	tpl *template.Template
)

func stringHandler(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "%s", printText)

	return nil
}

func htmlHandler(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tpl.Execute(w, struct {
		PrintText string
	}{
		PrintText: printText,
	})

	return nil
}

func jsonHandler(w http.ResponseWriter, r *http.Request) error {
	b, err := json.Marshal(struct {
		Message string `json:"message"`
	}{
		Message: fmt.Sprintf("%s", printText),
	})
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(b)

	return nil
}

func router(w http.ResponseWriter, r *http.Request) {
	sleepTime := time.Duration(30)

	var err error
	switch r.URL.Path {
	case rootPath:
		err = jsonHandler(w, r)
	case helloStringPath:
		err = stringHandler(w, r)
	case helloHTMLPath:
		err = htmlHandler(w, r)
	case helloJSONPath:
		err = jsonHandler(w, r)
	case slothHelloStringPath:
		time.Sleep(time.Second * sleepTime)
		err = stringHandler(w, r)
	case slothHelloHTMLPath:
		time.Sleep(time.Second * sleepTime)
		err = htmlHandler(w, r)
	case slothHelloJSONPath:
		time.Sleep(time.Second * sleepTime)
		err = jsonHandler(w, r)
	default:
		notFoundHandler(w, r)
		loggingAccess(r, http.StatusNotFound)
		return
	}

	if err != nil {
		log.Println(err)
		internalServerErrorHandler(w, r)
		loggingAccess(r, http.StatusInternalServerError)
		return
	}

	loggingAccess(r, http.StatusOK)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	errorHandler(w, r, http.StatusNotFound, "not found")
}

func internalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	errorHandler(w, r, http.StatusInternalServerError, "internal server error")
}

func errorHandler(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"message":"%s"}`, message)
}

func loggingAccess(r *http.Request, statusCode int) {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded == "" {
		forwarded = "-"
	}
	agent := r.UserAgent()
	if agent == "" {
		agent = "-"
	}
	referer := r.Referer()
	if referer == "" {
		referer = "-"
	}
	log.Printf("%s %s %s %s %s %d\n", r.RemoteAddr, forwarded, agent, referer, r.RequestURI, statusCode)
}

func main() {
	if txt := os.Getenv("PRINT_TEXT"); txt != "" {
		printText = txt
	}

	html := `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>{{.PrintText}}</title>
</head>
<body>
    <h1>{{.PrintText}}</h1>
</body>
</html>
`

	var err error
	if tpl, err = template.New("response").Parse(html); err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc(rootPath, router)
	http.HandleFunc(helloStringPath, router)
	http.HandleFunc(helloHTMLPath, router)
	http.HandleFunc(helloJSONPath, router)
	http.HandleFunc(slothHelloStringPath, router)
	http.HandleFunc(slothHelloHTMLPath, router)
	http.HandleFunc(slothHelloJSONPath, router)

	log.Println("start hello server.")
	http.ListenAndServe(":8080", nil)
}
