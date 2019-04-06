package option

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strings"
)

// Option command link options
type Option struct {
	ImportPath   string `json:importpath,omitempty`
	Modifier     string `json:modifier,omitempty`
	ExtraFlags   string `json:extraflags,omitempty`
	BuildRoot    string `json:buildroot,omitempty`
	BuildDir     string `json:builddir,omitempty`
	BuildPath    string `json:buildpath,omitempty`
	BuildContrib string `json:buildcontrib,omitempty`
	BuildSrc     string `json:buildsrc,omitempty`
	BuildBin     string `json:buildbin,omitempty`
	DestPath     string `json:destpath,omitempty`
}

// Fill fill up Option
func (opt *Option) Fill(args []string) {
	var importpath, modifier, extraflags string

	// loop the args to find the first with "-"
	idx := 0

	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			idx = i
			break
		}
	}

	// build the extraflags
	if idx > 0 {
		for _, arg := range args[idx:] {
			extraflags += arg + " "
		}
		args = args[:idx]
	}

	for i, arg := range args {
		if i == 0 {
			// split importpath from modifiers
			re := regexp.MustCompile(`(.*\/.*\/\w+)(.*)`)
			m := re.FindStringSubmatch(arg)
			if len(m) > 0 {
				importpath = m[1]
				modifier = m[2]
			} else {
				modifier = arg
			}
		} else {
			// "foo ..." equals to "foo..."
			re := regexp.MustCompile(`.*\w$`)
			if re.MatchString(modifier) && arg == "..." {
				modifier += arg
				// "foo bar ... baz" should be kept while bar in "foo ... bar" should be ignored
				if i == 1 {
					break
				}
			} else {
				modifier += " " + arg
			}
		}
	}

	if len(importpath) > 0 {
		opt.ImportPath = importpath
	}
	if len(modifier) > 0 {
		opt.Modifier = modifier
	}
	if len(extraflags) > 0 {
		opt.ExtraFlags = extraflags
	}
}

// Save save option to file
func (opt Option) Save() {
	f, _ := json.Marshal(&opt)
	ioutil.WriteFile("/tmp/golang.json", f, 0644)
}

// Load load options from file
func (opt *Option) Load() {
	if f, e := ioutil.ReadFile("/tmp/golang.json"); e == nil {
		json.Unmarshal(f, &opt)
	}
}
