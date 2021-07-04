package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/marguerite/go-stdlib/dir"
)

const (
	ARCH = runtime.GOARCH
)

var (
	HOME = os.Getenv("HOME") //+ "/test"
)

func ABI() string {
	return strings.Join(strings.Split(runtime.Version(), ".")[:2], ".")
}

func Version() string {
	return strings.TrimPrefix(ABI(), "go")
}

func GOPATH() string {
	return filepath.Join(HOME, "/rpmbuild/BUILD/go")
}

func GOSRC() string {
	return filepath.Join(GOPATH(), "/src")
}

func GOBIN() string {
	return filepath.Join(GOPATH(), "/bin")
}

func GOMOD() string {
	return filepath.Join(GOPATH(), "/pkg/mod")
}

func GOCONTRIBSRC() string {
	return filepath.Join("/usr/share/go/", Version(), "/contrib/src")
}

func RPM_BUILD_DIR() string {
	matches, err := dir.Glob(filepath.Join(HOME, "/rpmbuild/BUILD/*"))
	if err != nil {
		panic(err)
	}
	return matches[0]
}

func RPM_BUILD_ROOT() string {
	val := os.Getenv("RPM_BUILD_ROOT")
	if len(val) > 0 {
		return val
	}

	matches, err := dir.Glob(filepath.Join(HOME, "/rpmbuild/BUILDROOT/*"))
	if err != nil {
		panic(err)
	}
	return matches[0]
}

func SPECFILE() string {
	matches, err := dir.Glob(filepath.Join(HOME + "/rpmbuild/SOURCES/*.spec"))
	if err != nil {
		panic(err)
	}
	return matches[0]
}
