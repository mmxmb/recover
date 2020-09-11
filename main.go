package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
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
	rw  *RecoveryResponseWriter
}

func NewRecoverMiddleware(dev bool) *Recovery {
	rm := &Recovery{dev: dev}
	rm.mux = http.NewServeMux()
	rm.mux.HandleFunc("/panic/", panicDemo)
	rm.mux.HandleFunc("/panic-after/", panicAfterDemo)
	rm.mux.HandleFunc("/", hello)
	rm.rw = &RecoveryResponseWriter{}
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
	rec.rw.ResponseWriter = w
	h.ServeHTTP(rec.rw, r)
	rec.rw.flush()
}

type RecoveryResponseWriter struct {
	http.ResponseWriter
	writes [][]byte
	status int
}

func (rw *RecoveryResponseWriter) Write(b []byte) (int, error) {
	rw.writes = append(rw.writes, b)
	return len(b), nil
}

func (rw *RecoveryResponseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
}

func (rw *RecoveryResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter does not support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (rw *RecoveryResponseWriter) Flush() {
	flusher, ok := rw.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}
	flusher.Flush()
}

func (rw *RecoveryResponseWriter) flush() error {
	if rw.status != 0 {
		rw.ResponseWriter.WriteHeader(rw.status)
	}
	for _, write := range rw.writes {
		_, err := rw.ResponseWriter.Write(write)
		if err != nil {
			return err
		}
	}
	rw.writes = nil // clear writes buffer after response is sent
	return nil
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
