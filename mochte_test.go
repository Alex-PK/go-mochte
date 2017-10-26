package mochte

import (
	"testing"
	"net/http"
)

func TestBasics(t *testing.T) {
	s := New(t)
	defer s.Close()

	s.
	ListenOrdered().
		Add(NewHandler().
		Method(GET).
		Path("/").
		Status(200).
		Body("OK...").
		AssertIsCalledAtLeastNTimes(1),
	).Run()

	res, err := http.Get(s.Url() + "/")
	if err != nil {
		t.Error(err)
	}

	t.Logf("Result: %#v", res)
}
