# Mochte

An HTTP server mocking framework that allows you to verify assertions on the routes and possibly check the order in
which the routes are called.

## Status

Under heavy development.

APIs can change, documentation is bare minimal. Keeping this in mind, suggestions and contributions are welcome. :)

## Example

You just need to create a server, add some routes with their own assertions, then run the server and defer its closing.

You then use a normal `http.Client` to connect to the exposed URL and the server will verify the assertions.

```go
import "http://github.com/alex-pk/mochte"

func TestBasics(t *testing.T) {
	defer mochte.NewServerOn(t, ":49999").
		ListenOrdered().
		Add(mochte.NewRoute().
		Method(GET).
		Path("/").
		Status(200).
		Body("OK...").
		AssertIsCalledAtLeastNTimes(1),
	).Run().Close()

	res, err := http.Get("http://localhost:49999" + "/")
	if err != nil {
		t.Error(err)
	}

	t.Logf("Result: %#v", res)
}
```

