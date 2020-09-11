package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

func main() {
	dev := flag.Bool("dev", false, "")
	flag.Parse()
	rm := NewRecoverMiddleware(*dev)
	log.Fatal(http.ListenAndServe(":3000", rm))
}

type Recovery struct {
	mux *http.ServeMux
	dev bool
}

func NewRecoverMiddleware(dev bool) *Recovery {
	rm := &Recovery{dev: dev}
	rm.mux = http.NewServeMux()
	rm.mux.HandleFunc("/panic/", panicDemo)
	rm.mux.HandleFunc("/panic-after/", panicAfterDemo)
	rm.mux.HandleFunc("/", hello)
	return rm
}

func (rec *Recovery) recover(w http.ResponseWriter) {
	if err := recover(); err != nil {
		log.Printf("panic: %s\nstacktrace: %s\n", err, string(debug.Stack()))
		if rec.dev {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "<h1>panic: %s</h1><pre>%s</pre>", err, string(debug.Stack()))
		} else {
			http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		}
	}
}

func (rec *Recovery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer rec.recover(w)
	h, _ := rec.mux.Handler(r)
	h.ServeHTTP(w, r)
}

func panicDemo(w http.ResponseWriter, r *http.Request) {
	funcThatPanics()
}

func panicAfterDemo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello!</h1>")
	funcThatPanics()
}

func funcThatPanics() {
	panic("Oh no!")
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Hello!</h1>")
}
