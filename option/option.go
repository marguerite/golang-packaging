package option

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

// Option command link options
type Option struct {
	BuildMode    string   `json:"buildmode,omitempty"`
	ImportPath   string   `json:"importpath,omitempty"`
	Modifiers    []string `json:"modifier,omitempty"`
	ExtraFlags   []string `json:"extraflags,omitempty"`
	BuildRoot    string   `json:"buildroot,omitempty"`
	BuildDir     string   `json:"builddir,omitempty"`
	BuildPath    string   `json:"buildpath,omitempty"`
	BuildContrib string   `json:"buildcontrib,omitempty"`
	BuildSrc     string   `json:"buildsrc,omitempty"`
	BuildBin     string   `json:"buildbin,omitempty"`
	DestPath     string   `json:"destpath,omitempty"`
}

// Parse parse command line arguments to Option
func (opt *Option) Parse(args []string) {
	var importPath string
	opt.BuildMode, args = parseBuildMode(args)
	opt.ExtraFlags, args = parseExtraFlags(args)
	importPath, args = parseImportPath(args)
	if len(opt.ImportPath) == 0 {
		opt.ImportPath = importPath
	}
	opt.Modifiers = parseModifiers(args)
}

func parseModifiers(args []string) []string {
	// [foo] => foo
	// [foo...] => foo/...
	// [foo/...] => foo/...
	// [foo bar] => foo bar
	// [foo ... bar] => foo/... bar
	// [foo bar ... bar] => foo bar/...
	// [foo bar ... baz] => foo bar/... baz
	modifiers := []string{}
	m := map[string]struct{}{}
	for i, v := range args {
		if i > 0 && v == "..." {
			modifiers[i-1] = modifiers[i-1] + "/..."
		} else {
			if _, ok := m[v]; !ok {
				m[v] = struct{}{}
				modifiers = append(modifiers, v)
			}
		}
	}
	return modifiers
}

func parseImportPath(args []string) (string, []string) {
	// if [foo bar ... baz] it means importPath has been found
	// it has only two results:
	// importpath, []string{}
	// "", args (passthrough)
	if len(args) == 1 {
		// golang.org/x/mobile/test
		str := strings.Join(args, " ")
		re := regexp.MustCompile(`(^.*?\/\w+\/[^\.]+).*$`)
		if re.MatchString(str) {
			m := re.FindStringSubmatch(str)
			return m[1], []string{}
		}
	}
	return "", args
}

func parseExtraFlags(args []string) ([]string, []string) {
	extraFlags := []string{}
	newArgs := make([]string, len(args))
	copy(newArgs, args)
	re := regexp.MustCompile(`^-\w+flags$`)
	for i, v := range args {
		// xxflags, treat the item after it as its content
		if re.MatchString(v) {
			extraFlags = append(extraFlags, v)
			newArgs = removeFromSlice(newArgs, v)
			if i+1 < len(args) {
				s := "'" + args[i+1] + "'"
				extraFlags = append(extraFlags, s)
				newArgs = removeFromSlice(newArgs, args[i+1])
			}
		} else {
			// select the "-" but not select the content of xxflags
			if strings.HasPrefix(v, "-") && (i == 0 || !re.MatchString(args[i-1])) {
				extraFlags = append(extraFlags, v)
				newArgs = removeFromSlice(newArgs, v)
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					s := "'" + args[i+1] + "'"
					extraFlags = append(extraFlags, s)
					newArgs = removeFromSlice(newArgs, args[i+1])
				}
			}
		}
	}
	return extraFlags, newArgs
}

func parseBuildMode(args []string) (string, []string) {
	supportedBuildModes := map[string]struct{}{
		"default":   {},
		"archive":   {},
		"c-archive": {},
		"c-shared":  {},
		"shared":    {},
		"exe":       {},
		"pie":       {},
		"plugin":    {},
	}

	str := strings.Join(args, " ")
	re := regexp.MustCompile(`^.*(-buildmode=([^ ]+)).*$`)

	if re.MatchString(str) {
		m := re.FindStringSubmatch(str)
		if _, ok := supportedBuildModes[m[2]]; ok {
			buildMode := m[2]
			if buildMode == "exe" {
				fmt.Println("-buildmode=exe is not recommended, please consider using '-buildmode=pie'.")
			}
			if buildMode == "archive" {
				fmt.Println("Although supported, openSUSE doesn't recommend to install static build results for Go. Please consider repackaging")
			}
			return buildMode, removeFromSlice(args, m[1])
		} else {
			fmt.Errorf("%s is not a supported buildmode, supported: default, archive, c-archive, c-shared, shared, exe, pie, plugin.\n", m[2])
		}
	}
	// default to buildmode "pie"
	return "pie", args
}

func removeFromSlice(slice []string, s string) []string {
	for i, v := range slice {
		if v == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
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
