package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

func main() {
	rm := NewRecoverMiddleware()
	log.Fatal(http.ListenAndServe(":3000", rm))
}

type RecoverMiddleware struct {
	mux *http.ServeMux
}

func NewRecoverMiddleware() *RecoverMiddleware {
	rm := &RecoverMiddleware{}
	rm.mux = http.NewServeMux()
	rm.mux.HandleFunc("/panic/", panicDemo)
	rm.mux.HandleFunc("/panic-after/", panicAfterDemo)
	rm.mux.HandleFunc("/", hello)
	return rm
}

func (rm *RecoverMiddleware) recover(w http.ResponseWriter) {
	if r := recover(); r != nil {
		log.Printf("panic: %s\nstacktrace: %s\n", r, string(debug.Stack()))
	}
	http.Error(w, "Something went wrong...", http.StatusInternalServerError)
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
