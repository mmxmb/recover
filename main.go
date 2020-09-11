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

type RecoverMiddleware struct {
	mux *http.ServeMux
	dev bool
}

func NewRecoverMiddleware(dev bool) *RecoverMiddleware {
	rm := &RecoverMiddleware{dev: dev}
	rm.mux = http.NewServeMux()
	rm.mux.HandleFunc("/panic/", panicDemo)
	rm.mux.HandleFunc("/panic-after/", panicAfterDemo)
	rm.mux.HandleFunc("/", hello)
	return rm
}

func (rm *RecoverMiddleware) recover(w http.ResponseWriter) {
	if r := recover(); r != nil {
		log.Printf("panic: %s\nstacktrace: %s\n", r, string(debug.Stack()))
		if rm.dev {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "<h1>panic: %s</h1><pre>%s</pre>", r, string(debug.Stack()))
		} else {
			http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		}
	}
}

func (rm *RecoverMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer rm.recover(w)
	h, _ := rm.mux.Handler(r)
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
