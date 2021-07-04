package main

import (
	"fmt"
	"os"
	"path/filepath"
)

//https://golang.org/ref/mod#mod-commands
// IsModuleAware if golang is module aware
func IsModuleAware() bool {
	switch os.Getenv("GO111MODULE") {
	case "on", "":
		return true
	case "off":
		return false
	case "auto":
		wd, _ := os.Getwd()
		for {
			f, _ := os.Open(wd)
			names, _ := f.Readdirnames(-1)
			f.Close()
			for _, v := range names {
				if filepath.Base(v) == "go.mod" {
					return true
				}
			}
			wd1 := filepath.Dir(wd)
			if wd1 == wd {
				break
			}
			wd = wd1
		}
		return false
	default:
		fmt.Printf("unkown GO111MODULE=%s\n", os.Getenv("GO111MODULE"))
		os.Exit(1)
	}

	return true
}
