package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/marguerite/go-stdlib/open3"
)

func buildPackage(args []string) {
	err := readImportPath()
	if err != nil {
		panic(err)
	}

	// can be "/...", "..." or "@versionsuffix"
	var modifier string
	if len(args) > 0 {
		modifier = args[len(args)-1]
		switch {
		case modifier == "...":
		case modifier[0] == '/':
			modifier = modifier[1:]
		case modifier[0] == '@':
		default:
			modifier = "/" + modifier
		}

		args = args[:len(args)-1]
	}
	opts := []string{"build", "-v", "-p", "4", "-x"}

	// special handling of "-s" flag
	f, _ := strconv.ParseFloat(Version(), 64)
	if f < 1.10 {
		opts = append(opts, "-s")
	}

	// special handling for arch ppc64/riscv64
	if ARCH == "ppc64" || ARCH == "riscv64" {
		opts = append(opts, "-buildmode=pie")
	}

	for _, arg := range args {
		opts = append(opts, arg)
	}

	opts = append(opts, IMPORTPATH+modifier)

	fmt.Printf("GOPATH=%s:%s\n", GOPATH(), GOCONTRIBSRC())
	fmt.Printf("GOBIN=%s\n", GOBIN())
	fmt.Printf("Running /usr/bin/go %s\n", strings.Join(opts, " "))

	cmd := exec.Command("/usr/bin/go", opts...)
	stdoutbuf := bytes.NewBuffer([]byte{})
	stderrbuf := bytes.NewBuffer([]byte{})

	wt, err := open3.Popen3(cmd, RPM_BUILD_DIR(), func(stdin io.WriteCloser, stdout, stderr io.ReadCloser, wt open3.Wait_thr) error {
		stdin.Close()
		stdoutbuf.ReadFrom(stdout)
		stderrbuf.ReadFrom(stderr)
		return nil
	}, fmt.Sprintf("GOPATH=%s:%s", GOPATH(), GOCONTRIBSRC()), "GOBIN="+GOBIN(), "GOCACHE="+filepath.Join(HOME, ".cache/go-build"))

	fmt.Println(stdoutbuf.String())
	fmt.Println(stderrbuf.String())

	if err != nil {
		fmt.Println(err)
	}

	if wt.Value != 0 {
		os.Exit(1)
	}
}
