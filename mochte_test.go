package mochte

import (
	"testing"
	"log"
)

func TestBasics(t *testing.T) {
	ms := New(t)
	defer ms.Close()

	ms.Add(NewHandler().
		Method(GET).
		Path("/").
		Status(200).
		Body("OK..."),
	).Run()

	log.Printf("%#v", ms)
}
