package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
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

	_, err := template.ParseFS(os.DirFS(pathServe), "*.html")
	if err != nil {
		log.Println(err)
		flag.Usage()
		os.Exit(1)
	}

	http.HandleFunc(
		"/", func(w http.ResponseWriter, r *http.Request) {
			templates, err := template.ParseFS(os.DirFS(pathServe), "*.html")
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("<html><h1>Error</h1><p>" + err.Error() + "</p></html>"))
				return
			}
			if err := templates.ExecuteTemplate(
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
		},
	)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Println(err)
	}
}
