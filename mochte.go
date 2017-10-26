package mochte

import (
	"net/http"
	"testing"
	"time"
)

type Server struct {
	t           *testing.T
	url         string
	handlers    []*Handler
	handlerStep int
	listenMode  int
}

func New(t *testing.T, url string) Server {
	if url == "" {
		// TODO: generate URL
		generatedUrl := "localhost:49999"
		url = generatedUrl
	}

	return Server{
		t:           t,
		url:         url,
		handlers:    []*Handler{},
		handlerStep: 0,
		listenMode:  LISTEN_ANY,
	}
}

const (
	LISTEN_ORDERED = iota
	LISTEN_ANY
)

func (self Server) Add(h *Handler) Server {
	self.handlers = append(self.handlers, h)
	return self
}

func (self Server) ListenOrdered() Server {
	self.listenMode = LISTEN_ORDERED
	return self
}

func (self Server) ListenAny() Server {
	self.listenMode = LISTEN_ANY
	return self
}

func (self Server) Run() {
	go func() {
		if err := http.ListenAndServe(self.url, &self); err != nil {
			self.t.Errorf("Unable to start HTTP server: %s", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
}

func (self Server) route(req *http.Request) *Handler {
	if self.listenMode == LISTEN_ORDERED {
		handler := self.handlers[self.handlerStep]
		self.handlerStep += 1
		if handler.isHandling(req) {
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
