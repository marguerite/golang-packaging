package main

import "testing"

func TestBase(t *testing.T) {
	e := "golang.org/x/text"
	b := base("golang.org/x/text/encoding/charmap")
	if b != e {
		t.Errorf("test base() failed, expected %s, found %s\n", e, b)
	}

	e = "gopkg.in/check.v1"
	b = base(e)
	if b != e {
		t.Errorf("test base() failed, expected %s, found %s\n", e, b)
	}
}
