package main

import (
	"path/filepath"

	"github.com/marguerite/go-stdlib/dir"
	"github.com/marguerite/go-stdlib/fileutils"
)

func installPackage() {
	files, _ := dir.Glob(filepath.Join(GOBIN(), "/*"))
	for _, v := range files {
		fileutils.Copy(v, filepath.Join(RPM_BUILD_ROOT(), "/usr/bin/", filepath.Base(v)))
	}
}
