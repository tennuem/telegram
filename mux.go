package telegram

import (
	"fmt"
	"sync"
)

var DefaultServeMux = &ServeMux{commands: make(map[string]Handler)}

type HandlerFunc func(*ResponseWriter, *Request)

func (f HandlerFunc) Serve(w *ResponseWriter, r *Request) {
	f(w, r)
}

type ServeMux struct {
	mu       sync.Mutex
	commands map[string]Handler
}

func NewServeMux() *ServeMux {
	return &ServeMux{commands: make(map[string]Handler)}
}

func (mux *ServeMux) Handle(command string, handler Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	mux.commands[command] = handler
}

func (mux *ServeMux) HandleFunc(command string, handler func(*ResponseWriter, *Request)) {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	mux.commands[command] = HandlerFunc(handler)
}

func (mux *ServeMux) Handler(r *Request) (Handler, error) {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	command := r.Message.Text
	if !r.Message.IsCommand() {
		command = "*"
	}
	v, ok := mux.commands[command]
	if !ok {
		return nil, fmt.Errorf("command: %s not found", command)
	}
	return v, nil
}

func (mux *ServeMux) Serve(w *ResponseWriter, r *Request) {
	h, err := mux.Handler(r)
	if err != nil {
		w.Text = err.Error()
		return
	}
	h.Serve(w, r)
}

func HandleFunc(pattern string, handler func(*ResponseWriter, *Request)) {
	DefaultServeMux.HandleFunc(pattern, HandlerFunc(handler))
}

func Handle(pattern string, handler Handler) {
	DefaultServeMux.Handle(pattern, handler)
}
