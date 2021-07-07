package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/mod/module"
)

func parseImportPath(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("no importpath given")
	}

	if len(args) > 2 {
		return "", errors.New("unknown input")
	}

	var path string
	if len(args) == 1 {
		path = args[0]
	} else {
		if args[1] == "..." {
			path = strings.Join(args, "/")
		}
	}

	// go help packages
	if err := module.CheckPath(path); err == nil {
		return path, nil
	}

	return "", errors.New("not a valid importpath")
}

func storeImportPath(path string) error {
	n, err := IMPORTPATH.WriteString(path)
	if err != nil {
		return err
	}
	if n != len(path) {
		return fmt.Errorf("importpath not fully written to %s\n", IMPORTPATH.Name())
	}
	return nil
}

func readImportPath() (string, error) {
	b, err := readLine(IMPORTPATH)
	if err != nil && err != io.EOF {
		return "", err
	}

	return string(b), nil
}
