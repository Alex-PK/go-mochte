package mochte

import (
	"net/http"
	"strings"
	"testing"
)

type Handler struct {
	method      string
	path        string
	status      int
	contentType string
	body        *string
	bodyFn      func() string

	callCount int

	runtimeChecks []checkFn
	finalChecks   []checkFn
}

type checkFn func(*testing.T)

func NewHandler() *Handler {
	return &Handler{
		method:      GET,
		path:        "/",
		status:      200,
		contentType: HTML,
		bodyFn:      func() string { return "" },

		callCount: 0,

		runtimeChecks: []checkFn{},
		finalChecks:   []checkFn{},
	}
}

func (self *Handler) Method(method string) *Handler {
	self.method = method
	return self
}

func (self *Handler) Get(path string) *Handler {
	return self.Method(GET).Path(path)
}

func (self *Handler) Head(path string) *Handler {
	return self.Method(HEAD).Path(path)
}

func (self *Handler) Post(path string) *Handler {
	return self.Method(POST).Path(path)
}

func (self *Handler) Put(path string) *Handler {
	return self.Method(PUT).Path(path)
}

func (self *Handler) Delete(path string) *Handler {
	return self.Method(DELETE).Path(path)
}

func (self *Handler) Path(path string) *Handler {
	self.path = path
	return self
}

func (self *Handler) Status(status int) *Handler {
	self.status = status
	return self
}

func (self *Handler) ContentType(contentType string) *Handler {
	self.contentType = contentType
	return self
}

func (self *Handler) Body(body string) *Handler {
	self.body = &body
	return self
}

func (self *Handler) BodyFn(f func() string) *Handler {
	self.bodyFn = f
	return self
}

func (self *Handler) AssertIsCalledNTimes(n int) *Handler {
	self.finalChecks = append(self.finalChecks, func(t *testing.T) {
		if n != self.callCount {
			t.Errorf("Expecting handler to be called %d times, called %d times", n, self.callCount)
		}
	})
	return self
}

func (self *Handler) AssertIsCalledAtLeastNTimes(n int) *Handler {
	self.finalChecks = append(self.finalChecks, func(t *testing.T) {
		if self.callCount >= n {
			t.Errorf("Expecting handler to be called at least %d times, called %d times", n, self.callCount)
		}
	})
	return self
}

func (self *Handler) AssertIsCalledNoMoreThanNTimes(n int) *Handler {
	self.finalChecks = append(self.finalChecks, func(t *testing.T) {
		if self.callCount < n {
			t.Errorf("Expecting handler to be called at least %d times, called %d times", n, self.callCount)
		}
	})
	return self
}

/*
 * 	Server-called methods
 */

func (self *Handler) isHandling(req *http.Request) bool {
	if req.Method != self.method {
		return false
	}
	if !strings.HasPrefix(req.URL.Path, self.path) {
		return false
	}

	// TODO: improve checks
	return true
}

func (self *Handler) handle(t *testing.T, w http.ResponseWriter, req *http.Request) {
	self.incCounter()

	for _, check := range self.runtimeChecks {
		check(t)
	}

	w.Header().Add("Content-type", self.getContentType())
	w.WriteHeader(self.getStatus())
	w.Write([]byte(self.getBody()))
}

func (self *Handler) runFinalChecks(t *testing.T) {
	for _, check := range self.finalChecks {
		check(t)
	}
}

func (self *Handler) getStatus() int {
	return self.status
}

func (self *Handler) getContentType() string {
	return self.contentType
}

func (self *Handler) getBody() string {
	if self.body != nil {
		return *self.body
	} else {
		return self.bodyFn()
	}
}

func (self *Handler) incCounter() {
	self.callCount += 1
}

/*
 * 	Getters
 */

func (self *Handler) GetCallCount() int {
	return self.callCount
}
