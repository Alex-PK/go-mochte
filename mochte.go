package mochte

import (
	"net/http"
	"testing"
	"time"
	"context"
)

// Server defines the mock HTTP server
type Server struct {
	t          *testing.T
	srv        *http.Server
	addr       string
	routes     []*Route
	routeStep  int
	listenMode int
}

// NewServer creates a new test server on a random port on localhost.
func NewServer(t *testing.T) *Server {
	return NewServerOn(t, ":49999") // TODO: generate "randomly"
}

// NewServerOn creates a new test server on an address you provide
func NewServerOn(t *testing.T, addr string) *Server {
	server := &Server{
		t:          t,
		addr:       addr,
		routes:     []*Route{},
		routeStep:  0,
		listenMode: listenAny,
	}

	server.srv = &http.Server{Addr: server.addr, Handler: server}

	return server
}

const (
	listenOrdered = iota
	listenAny
)

// URL returns a URL that can be used by an HTTP Client to connect to the server
func (server *Server) URL() string {
	return "http://" + server.addr
}

// Add allows to add a route handler to the list of tested routes
func (server *Server) Add(h *Route) *Server {
	server.routes = append(server.routes, h)
	return server
}

// ListenOrdered sets the server in "ordered mode". The added routes must be called in order by the client,
// or make the test fail.
//
// Each route can also have its own assertions
func (server *Server) ListenOrdered() *Server {
	server.listenMode = listenOrdered
	return server
}

// ListenAny sets the server in "any order mode". The routes can be definied in any order, and the server does not
// expect them to be called in a specific order.
//
// Each route should have its own assertions.
func (server *Server) ListenAny() *Server {
	server.listenMode = listenAny
	return server
}

// Run spawns a goroutine containing the server itself, effectively starting to listen for connections.
//
// This function must be called before making connections to the server, for the tests to work
func (server *Server) Run() *Server {
	go func() {
		if err := server.srv.ListenAndServe(); err != nil {
			server.t.Log(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	return server
}

// Close shuts down the server and runs the final checks on all the routes.
//
// This function must be called at the end of the test case (or with defer) to turn off the server and
// run the final tests.
func (server *Server) Close() {
	server.t.Log("Shutting down server")
	err := server.srv.Shutdown(context.Background())
	if err != nil {
		server.t.Fatal("Failed to shutdown the server")
	}

	for _, route := range server.routes {
		route.runFinalChecks(server.t)
	}
}

/*
 *	Internally used functions
 */

func (server *Server) route(req *http.Request) *Route {
	if server.listenMode == listenOrdered {
		route := server.routes[server.routeStep]
		server.routeStep++
		if route != nil && route.isHandling(req) {
			return route
		}
	} else {
		for _, route := range server.routes {
			if route.isHandling(req) {
				return route
			}
		}
	}

	return nil
}

func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	route := server.route(req)
	if route == nil {
		server.t.Errorf("Unexpected endpoint called: %s %s", req.Method, req.RequestURI)
		return
	}

	route.handle(server.t, w, req)
}
