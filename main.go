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

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: golang-packaging [action] [option list]")
		os.Exit(1)
	}

	importpathFile := filepath.Join(RPM_BUILD_DIR(), "importpath.txt")
	if _, err := os.Stat(importpathFile); os.IsNotExist(err) {
		IMPORTPATH, _ = os.Create(importpathFile)
	} else {
		IMPORTPATH, _ = os.Open(importpathFile)
	}
	defer IMPORTPATH.Close()

	module := IsModuleAware()

	switch os.Args[1] {
	case "prep":
		BUILDREQUIRES = getBuildRequires(SPECFILE())
		VENDORED = parseModulesTxt(filepath.Join(RPM_BUILD_DIR(), "vendor/modules.txt"))
		prepareBuildEnvironment(os.Args[2:], module)
	case "build":
		buildPackage(os.Args[2:])
	case "install":
		installPackage()
	default:
		fmt.Printf("unknown action %s\n", os.Args[1])
		os.Exit(1)
	}
}
