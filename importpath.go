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

func readImportPath(path string) (string, error) {
	b := make([]byte, 0, 50)

	for {
		b1 := make([]byte, 0, 1)
		n, err := IMPORTPATH.Read(b1)
		if err == io.EOF {
			break
		}
		if n == 0 {
			break
		}
		if n > 1 {
			return "", errors.New("read more than 1 byte")
		}
		b = append(b, b1[0])
	}

	return string(b), nil
}
