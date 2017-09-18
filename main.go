package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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
	printText       = "Hello World!"
	shutdownTimeout = 30

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
	if st := os.Getenv("SHUTDOWN_TIMEOUT"); st != "" {
		var err error
		if shutdownTimeout, err = strconv.Atoi(st); err != nil {
			log.Println("value in SHUTDOWN_TIMEOUT is invalid")
			os.Exit(1)
		}
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
		log.Println(err)
		os.Exit(1)
	}

	os.Exit(run())
}

func run() int {
	mux := http.NewServeMux()

	mux.HandleFunc(rootPath, router)
	mux.HandleFunc(helloStringPath, router)
	mux.HandleFunc(helloHTMLPath, router)
	mux.HandleFunc(helloJSONPath, router)
	mux.HandleFunc(slothHelloStringPath, router)
	mux.HandleFunc(slothHelloHTMLPath, router)
	mux.HandleFunc(slothHelloJSONPath, router)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	cerr := make(chan error, 1)

	go func() {
		cerr <- srv.ListenAndServe()
	}()

	log.Println("running hello server.")

	select {
	case err := <-cerr:
		if err != nil && err != http.ErrServerClosed {
			log.Println(err)
			return 1
		}
		return 0
	case <-waitSginal():
	}

	log.Println("stopping hello server.")
	defer log.Println("stopped hello server.")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(shutdownTimeout))
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		srv.Close()
		log.Println(err)
		return 1
	}

	return 0
}

func waitSginal() <-chan struct{} {
	ret := make(chan struct{}, 1)

	quit := make(chan os.Signal, 1)
	sigs := []os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	}
	signal.Notify(quit, sigs...)

	go func() {
		<-quit
		ret <- struct{}{}
	}()

	return ret
}
