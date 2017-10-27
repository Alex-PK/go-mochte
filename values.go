package mochte

const (
	// GET is a const alias fopr the corresponding HTTP method
	GET    = "GET"

	// POST is a const alias fopr the corresponding HTTP method
	POST   = "POST"

	// HEAD is a const alias fopr the corresponding HTTP method
	HEAD   = "HEAD"

	// PUT is a const alias fopr the corresponding HTTP method
	PUT    = "PUT"

	// DELETE is a const alias fopr the corresponding HTTP method
	DELETE = "DELETE"
)

const (
	// HTML is a utility const for the corresponding ContentType
	HTML = "text/html"

	// JSON is a utility const for the corresponding ContentType
	JSON = "application/json"
)

const (
	// DebugTrace allows to trace execution by logging calls even when there is no failure
	DebugTrace = 1 << iota

	// DebugHeaders allows to dump request headers on every call
	DebugHeaders

	// DebugBody allows to dump the request body
	DebugBody

	// DebugNone disables tracing and debugging messages during assertions. Do not mix with other levels
	DebugNone = 0
)

const (
	listenOrdered = iota
	listenAny
)
