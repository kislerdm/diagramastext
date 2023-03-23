package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
)

type env struct {
	API_URL string
	VERSION string
	TOKEN   string
}

type impute struct {
	Env env
}

func main() {
	var (
		port      string
		pathServe string
	)
	flag.StringVar(&port, "port", "8080", "port to listen on")
	flag.StringVar(&pathServe, "path", "/public", "path to index.html file")
	flag.Parse()

	h, err := newHandler(pathServe)
	if err != nil {
		log.Println(err)
		flag.Usage()
		os.Exit(1)
	}

	if err := http.ListenAndServe(":"+port, h); err != nil {
		log.Println(err)
	}
}

func newHandler(p string) (http.Handler, error) {
	if _, err := template.ParseFiles(path.Join(p, "index.html")); err != nil {
		return nil, err
	}
	return handler{
		Dir:          p,
		filesHandler: http.FileServer(http.Dir(p)),
	}, nil
}

type handler struct {
	Dir string

	filesHandler http.Handler
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		t, err := template.ParseFiles(path.Join(h.Dir, "index.html"))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("<html><h1>Error</h1><p>" + err.Error() + "</p></html>"))
			return
		}
		if err := t.ExecuteTemplate(
			w, "index.html", impute{
				Env: env{
					API_URL: os.Getenv("API_URL"),
					VERSION: os.Getenv("VERSION"),
					TOKEN:   os.Getenv("TOKEN"),
				},
			},
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("<html><h1>Error</h1><p>" + err.Error() + "</p></html>"))
			return
		}
	default:
		h.filesHandler.ServeHTTP(w, r)
	}
}
