package mochte

import (
	"net/http"
	"testing"
	"time"
	"context"
)

type Server struct {
	t           *testing.T
	srv         *http.Server
	url         string
	handlers    []*Handler
	handlerStep int
	listenMode  int
}

func New(t *testing.T) *Server {

	self := &Server{
		t:           t,
		url:         ":49999", // TODO
		handlers:    []*Handler{},
		handlerStep: 0,
		listenMode:  LISTEN_ANY,
	}

	self.srv = &http.Server{Addr: self.url, Handler: self}

	return self
}

const (
	LISTEN_ORDERED = iota
	LISTEN_ANY
)

func (self *Server) Url() string {
	return self.url
}

func (self *Server) Add(h *Handler) *Server {
	self.handlers = append(self.handlers, h)
	return self
}

func (self *Server) ListenOrdered() *Server {
	self.listenMode = LISTEN_ORDERED
	return self
}

func (self *Server) ListenAny() *Server {
	self.listenMode = LISTEN_ANY
	return self
}

func (self *Server) Run() *Server {

	go func() {
		if err := self.srv.ListenAndServe(); err != nil {
			self.t.Errorf("Unable to start HTTP server: %s", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	return self
}

func (self *Server) Close() {
	self.t.Log("Shutting down server")
	self.srv.Shutdown(context.Background())
}

func (self *Server) route(req *http.Request) *Handler {
	if self.listenMode == LISTEN_ORDERED {
		handler := self.handlers[self.handlerStep]
		self.handlerStep += 1
		if handler != nil && handler.isHandling(req) {
			return handler
		}
	} else {
		for _, handler := range self.handlers {
			if handler.isHandling(req) {
				return handler
			}
		}
	}

	return nil
}

func (self *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := self.route(req)
	if handler == nil {
		self.t.Errorf("Unexpected endpoint called: %s %s", req.Method, req.RequestURI)
		return
	}

	handler.incCounter()
	w.Header().Add("Content-type", handler.getContentType())
	w.WriteHeader(handler.getStatus())
	w.Write([]byte(handler.getBody()))
}
