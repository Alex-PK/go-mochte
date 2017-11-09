package mochte

import (
	"net/http"
	"strings"
	"testing"
	"io/ioutil"
)

// Route contains the definition of a route that can be added to the Server
type Route struct {
	method      string
	path        string
	status      int
	contentType string
	body        *string
	bodyFn      func() string

	callCount int

	failed        checkResult
	runtimeChecks []reqCheckFn
	finalChecks   []simpleCheckFn
}

type checkResult bool

const (
	checkOK   checkResult = false
	checkFAIL checkResult = true
)

type simpleCheckFn func(*testing.T) checkResult
type reqCheckFn func(*testing.T, *http.Request) checkResult

// NewRoute creates a new route to be added to the server.
//
// By default it will listen to GET requests to the / path. Use the builder methods to configure it.
func NewRoute() *Route {
	return &Route{
		method:      GET,
		path:        "/",
		status:      200,
		contentType: HTML,
		bodyFn:      func() string { return "" },

		callCount: 0,

		runtimeChecks: []reqCheckFn{},
		finalChecks:   []simpleCheckFn{},
	}
}

// Method allows to define which method the route will listen to.
//
// You can use constants for some standard methods (GET, POST, PUT, HEAD, DELETE), or specify your own (uppercase).
func (route *Route) Method(method string) *Route {
	route.method = method
	return route
}

// Get is a shortcut to specify both the GET method and a Path to listen on
func (route *Route) Get(path string) *Route {
	return route.Method(GET).Path(path)
}

// Head is a shortcut to specify both the HEAD method and a Path to listen on
func (route *Route) Head(path string) *Route {
	return route.Method(HEAD).Path(path)
}

// Post is a shortcut to specify both the POST method and a Path to listen on
func (route *Route) Post(path string) *Route {
	return route.Method(POST).Path(path)
}

// Put is a shortcut to specify both the PUT method and a Path to listen on
func (route *Route) Put(path string) *Route {
	return route.Method(PUT).Path(path)
}

// Delete is a shortcut to specify both the DELETE method and a Path to listen on
func (route *Route) Delete(path string) *Route {
	return route.Method(DELETE).Path(path)
}

// Path allows to define the path the route will be listening on.
//
// It is treated as the beginning of the path, so / is a catch-all, and should be defined last
// when using ListenAny
//
// TODO: implement regexes or a simpler method
func (route *Route) Path(path string) *Route {
	route.path = path
	return route
}

// Status allows to define the status code that will be returned by the route.
//
// Default: 200
func (route *Route) Status(status int) *Route {
	route.status = status
	return route
}

// ContentType allows to define the Content-type header that will be returned in the response.
//
// Default: text/html
func (route *Route) ContentType(contentType string) *Route {
	route.contentType = contentType
	return route
}

// Body allows to define the body string returned in the response.
//
// If you need to generate a dynamic body, see the BodyFn method.
//
// If both Body and BodyFn are passed, Body has the highest priority, and BodyFn is ignored.
func (route *Route) Body(body string) *Route {
	route.body = &body
	return route
}

// BodyFn allows to pass a function to genrate the body of the response
//
// If you need to generate a static body, see the Body method.
//
// If both Body and BodyFn are passed, Body has the highest priority, and BodyFn is ignored.
func (route *Route) BodyFn(f func() string) *Route {
	route.bodyFn = f
	return route
}

// AssertIsCalledNTimes adds a (final) assertion checking that this route is called exactly n times during the test.
func (route *Route) AssertIsCalledNTimes(n int) *Route {
	route.finalChecks = append(route.finalChecks, func(t *testing.T) checkResult {
		if n != route.callCount {
			t.Logf("Expecting handler to be called %d time(s), called %d time(s)", n, route.callCount)
			return checkFAIL
		}
		return checkOK
	})
	return route
}

// AssertIsCalledAtLeastNTimes adds a (final) assertion checking that this route is called at least n times during the test.
func (route *Route) AssertIsCalledAtLeastNTimes(n int) *Route {
	route.finalChecks = append(route.finalChecks, func(t *testing.T) checkResult {
		if route.callCount < n {
			t.Logf("Expecting handler to be called at least %d time(s), called %d time(s)", n, route.callCount)
			return checkFAIL
		}
		return checkOK
	})
	return route
}

// AssertIsCalledNoMoreThanNTimes adds a (final) assertion checking that this route is called no more than n times during the test.
func (route *Route) AssertIsCalledNoMoreThanNTimes(n int) *Route {
	route.finalChecks = append(route.finalChecks, func(t *testing.T) checkResult {
		if route.callCount > n {
			t.Logf("Expecting handler to be called at least %d time(s), called %d time(s)", n, route.callCount)
			return checkFAIL
		}
		return checkOK
	})
	return route
}

// AssertHasContentType adds a check for the Content-type during request processing
func (route *Route) AssertHasContentType(ct string) *Route {
	route.runtimeChecks = append(route.runtimeChecks, func(t *testing.T, req *http.Request) checkResult {
		contentType := req.Header.Get("Content-type")
		if contentType == ct {
			return checkOK
		}
		return checkFAIL
	})
	return route
}

// AssertHasContentType allows to add a custom function to check he body of the request
func (route *Route) AssertWithBody(f func(t *testing.T, body []byte) bool) *Route {
	route.runtimeChecks = append(route.runtimeChecks, func(t *testing.T, req *http.Request) checkResult {
		reqBody, _ := req.GetBody()
		body, err := ioutil.ReadAll(reqBody)
		if err != nil {
			t.Error("Unable to read body from request: %s", err)
			return checkFAIL
		}

		if f(t, body) {
			return checkOK
		}

		return checkFAIL
	})
	return route
}

/*
 * 	Server-called methods
 */

func (route *Route) isHandling(req *http.Request) bool {
	if req.Method != route.method {
		return false
	}
	if !strings.HasPrefix(req.URL.Path, route.path) {
		return false
	}

	// TODO: improve checks
	return true
}

func (route *Route) handle(t *testing.T, w http.ResponseWriter, req *http.Request, debugMode int) {
	route.incCounter()

	if debugMode&DebugTrace > 0 {
		t.Logf("Running %d request checks on %s %s", len(route.runtimeChecks), route.method, route.path)
	}

	for _, check := range route.runtimeChecks {
		route.failed = check(t, req) || route.failed
	}

	w.Header().Add("Content-type", route.getContentType())
	w.WriteHeader(route.getStatus())
	w.Write([]byte(route.getBody()))
}

func (route *Route) runFinalChecks(t *testing.T, debugMode int) {
	if debugMode&DebugTrace > 0 {
		t.Logf("Running %d final checks on %s %s", len(route.finalChecks), route.method, route.path)
	}

	for _, check := range route.finalChecks {
		route.failed = check(t) || route.failed
	}

	if route.failed {
		t.Fail()
	}
}

func (route *Route) getStatus() int {
	return route.status
}

func (route *Route) getContentType() string {
	return route.contentType
}

func (route *Route) getBody() string {
	if route.body != nil {
		return *route.body
	} else if route.bodyFn != nil {
		return route.bodyFn()
	}

	return ""
}

func (route *Route) incCounter() {
	route.callCount++
}

/*
 * 	Getters
 */

// GetCallCount returns the number of times the route was hit during the test.
//
// This function should be rarely used, as setting an assertion is the best way to check this.
func (route *Route) GetCallCount() int {
	return route.callCount
}
