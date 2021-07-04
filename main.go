package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	IMPORTPATH    *os.File
	BUILDREQUIRES map[string]string
	VENDORED      map[string]moduleline
)

func init() {
	importpathFile := filepath.Join(RPM_BUILD_DIR(), "importpath.txt")
	if _, err := os.Stat(importpathFile); os.IsNotExist(err) {
		IMPORTPATH, _ = os.Create(importpathFile)
	} else {
		IMPORTPATH, _ = os.Open(importpathFile)
	}

	BUILDREQUIRES = getBuildRequires(SPECFILE())

	VENDORED = parseModulesTxt(filepath.Join(RPM_BUILD_DIR(), "vendor/modules.txt"))
}

func main() {
	defer IMPORTPATH.Close()

	if len(os.Args) == 1 {
		fmt.Println("usage: golang-packaging [action] [option list]")
		os.Exit(1)
	}

	module := IsModuleAware()

	switch os.Args[1] {
	case "prep":
		prepareBuildEnvironment(os.Args[2:], module)
	default:
		fmt.Printf("unknown action %s\n", os.Args[1])
		os.Exit(1)
	}
}
