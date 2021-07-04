package main

import (
	"testing"
)

func TestParseImportPath(t *testing.T) {
	s := []string{"github.com/marguerite/golang-packaging", "..."}

	path, err := parseImportPath(s)
	if path != "github.com/marguerite/golang-packaging/..." || err != nil {
		t.Errorf("parseImportPath failed, expected github.com/marguerite/golang-packaging, got %s", path)
	}
}
