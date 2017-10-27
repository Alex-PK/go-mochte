package mochte

import (
	"net/http"
	"testing"
)

func TestBasics(t *testing.T) {
	defer NewServerOn(t, ":49999").
		ListenOrdered().
		Add(NewRoute().
			Method(GET).
			Path("/").
			Status(200).
			Body("OK...").
			AssertIsCalledAtLeastNTimes(1),
		).
		Run().Close()

	res, err := http.Get("http://localhost:49999" + "/")
	if err != nil {
		t.Error(err)
	}

	t.Logf("Result: %#v", res)
}
