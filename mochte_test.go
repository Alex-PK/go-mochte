package mochte

import (
	"testing"
	"log"
)

func TestBasics(t *testing.T) {
	ms := New(t, "")

	ms.Add(NewHandler().
		Method(GET).
		Path("/").
		Status(200).
		Body("OK..."),
	)

	log.Printf("%#v", ms)
}
