package main

import (
	"flag"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
)

type env struct {
	ApiURL  string
	VERSION string
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
	if _, err := template.ParseFiles(path.Join(p, "src", "index.js")); err != nil {
		return nil, err
	}
	return handler{
		Dir:          p,
		filesHandler: http.FileServer(http.Dir(p)),
		imputation: impute{
			Env: env{
				ApiURL:  os.Getenv("API_URL"),
				VERSION: os.Getenv("VERSION"),
			},
		},
	}, nil
}

type handler struct {
	Dir        string
	imputation impute

	filesHandler http.Handler
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	w.Header().Add("Content-Type", mime.TypeByExtension(path.Ext(p)))

	switch p {
	case "/":
		t, err := template.ParseFiles(path.Join(h.Dir, "index.html"))
		w.Header().Set("Content-Type", mime.TypeByExtension("html"))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("<html><h1>Error</h1><p>" + err.Error() + "</p></html>"))
			return
		}
		if err := t.ExecuteTemplate(w, "index.html", h.imputation); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("<html><h1>Error</h1><p>" + err.Error() + "</p></html>"))
			return
		}
	case "/src/config.ts":
		t, err := template.ParseFiles(path.Join(h.Dir, "src", "config.ts"))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := t.ExecuteTemplate(w, "config.ts", h.imputation); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		h.filesHandler.ServeHTTP(w, r)
	}
}
