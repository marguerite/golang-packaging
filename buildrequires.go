package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/marguerite/diagnose/zypp/search"
	"github.com/openSUSE-zh/specfile"
)

func getBuildRequires(s string) map[string]string {
	f, err := os.Open(s)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	parser, err := specfile.NewParser(f)
	if err != nil {
		panic(err)
	}
	parser.Parse()
	buildrequires := parser.Spec.FindDependencies("BuildRequires")
	dependencies := make(map[string]string)
	for _, v := range buildrequires {
		var name, version string
		switch {
		case v == "golang-packaging":
		case strings.HasPrefix(v, "golang-"):
			name = rpmName2Importpath(v)
			version = "v" + fixVersion(name, findVersion(v))
		case strings.HasPrefix(v, "golang("):
			name, version = findNameVersion(v)
		default:
		}

		if _, ok := dependencies[name]; len(name) > 0 && !ok {
			dependencies[name] = version
		}
	}
	return dependencies
}

// convert rpm name to importpath
func rpmName2Importpath(s string) string {
	// golang-github-linuxdeepin-go-lib
	// golang-org-x-tools
	s = strings.TrimPrefix(s, "golang-")
	var host, author string
	for i := 0; i < 2; i++ {
		idx := strings.Index(s, "-")
		switch i {
		case 0:
			switch s[:idx] {
			case "org":
				host = "golang.org"
			default:
				host = s[:idx] + ".com"
			}
		case 1:
			author = s[:idx]
		}
		s = s[idx+1:]
	}
	return host + "/" + author + "/" + strings.TrimPrefix(s, "-")
}

// fixVersion golang supports v0 or v1 only, if > 2, the importpath/package name should be suffixed
func fixVersion(name, version string) string {
	if strings.Index(filepath.Base(name), ".") > 0 {
		return version
	}

	return "0." + version[strings.Index(version, ".")+1:]
}

func findVersion(s string) string {
	out, err := exec.Command("/usr/bin/rpm", "-q", "--qf", "%{VERSION}", s).Output()
	if err != nil {
		fmt.Printf("required package %s not installed.\n", s)
		os.Exit(1)
	}

	return string(out)
}

func findNameVersion(s string) (string, string) {
	searchables := search.NewSearch(s, false, "--provides")
	// golang(pkkg.deepin.io/lib/strv)
	s = s[7 : len(s)-1]
	if len(searchables) > 0 {
		return s, "v" + fixVersion(s, searchables[0].Version[:strings.Index(searchables[0].Version, "-")])
	}
	return "", ""
}
