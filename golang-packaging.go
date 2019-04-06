package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func copyDir(src, dst string) {
	log.Printf("Copying all files and directories from %s to %s...", src, dst)

	if _, e := os.Stat(dst); e != nil {
		log.Printf("%s doesn't exist, making...", dst)
		os.MkdirAll(dst, 0755)
	}

	sd, e := os.Open(src)
	if e != nil {
		log.Fatalf("%s doesn't exist, making...")
	}

	files, e := sd.Readdirnames(-1)
	if e != nil {
		log.Fatalf("Failed to read all files and directories for %s, read %v.", src, files)
	}

	for _, f := range files {
		filepath.Walk(f, func(path string, info os.FileInfo, err error) error {
			file := filepath.Base(path)
			if info.Mode().IsRegular() {
				copyFile(path, filepath.Join(dst, file))
			}
			if info.Mode().IsDir() {
				copyDir(path, filepath.Join(dst, file))
			}
			if info.Mode()&os.ModeSymlink != 0 {
				symlink, e := os.Readlink(path)
				if e != nil {
					log.Fatalf("Could not follow %s's symlink", path)
				}

				// make absolute symlink
				if !filepath.IsAbs(symlink) {
					log.Printf("Non-absolute symlink found: %s", symlink)
					log.Println("Converting to absolute path")
					symlink = filepath.Join(filepath.Dir(path), symlink)
				}

				// non existent symlink target
				if _, e := os.Stat(symlink); os.IsNotExist(e) {
					log.Fatalf("Non existent symlink target found: %s, quit.", symlink)
				}

				e = os.Remove(path)
				if e != nil {
					log.Fatalf("Failed to remove %s", path)
				}

				copyFile(symlink, filepath.Join(dst, file))
			}
			return nil
		})
	}
}

func copyFile(src, dst string) {
	sd, e := os.Open(src)
	if e != nil {
		log.Fatalf("Failed to open %s file descriptor", src)
	}
	defer sd.Close()

	fd, e := os.Create(dst)
	if e != nil {
		log.Fatalf("Failed to create %s file descriptor", dst)
	}
	defer fd.Close()

	_, e = io.Copy(fd, sd)
	if e != nil {
		log.Fatalf("Failed to copy %s to %s", src, dst)
	}

	mode, _ := os.Stat(src)
	e = os.Chmod(dst, mode.Mode())
	if e != nil {
		log.Fatalf("Failed to sync file permissions from %s to %s", src, dst)
	}
}

func loadArg(arg string) string {
	supportedArgs := map[string]struct{}{"importpath": {}, "modifier": {}, "extraflags": {}}

	if _, ok := supportedArgs[arg]; !ok {
		log.Fatal("Only support: importpath, modifier, extraflags")
	}

	argFile := filepath.Join("/tmp", arg)

	if _, e := os.Stat(argFile); os.IsNotExist(e) {
		log.Fatalf("Failed to read %s. Is it there?", argFile)
	}

	fd, e := os.OpenFile(argFile, os.O_RDWR, 0644)
	if e != nil {
		log.Fatalf("Failed to open %s, please check its permission.", argFile)
	}
	defer fd.Close()

	b, e := ioutil.ReadAll(fd)
	if e != nil {
		log.Fatalf("Could not read content of %s", argFile)
	}

	return string(b)
}

func storeArg(arg string, content string) {
	path := filepath.Join("/tmp", arg)
	// create it if doesn't exist
	if _, e := os.Stat(path); os.IsNotExist(e) {
		fd, e := os.Create(path)
		if e != nil {
			log.Fatalf("Could not create %s", path)
		}
		defer fd.Close()
		fd.WriteString(content)
	}
}

// Option command link options
type Option struct {
	Importpath string
	Modifier   string
	Extraflags string
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
		opt.Importpath = importpath
	}
	if len(modifier) > 0 {
		opt.Modifier = modifier
	}
	if len(extraflags) > 0 {
		opt.Extraflags = extraflags
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

func goBuild(command string, options []string, path string) {
	flags := append(append([]string{command}, options...), []string{path}...)

	var outBuf, errBuf bytes.Buffer
	var errOut, errErr error

	cmd := exec.Command("/usr/bin/go", flags...)
	env := append(os.Environ(), "GOPATH="+buildPath()+":"+buildContrib())
	env = append(env, "GOBIN="+buildBin())
	cmd.Env = env

	outIn, _ := cmd.StdoutPipe()
	errIn, _ := cmd.StderrPipe()

	out := io.MultiWriter(os.Stdout, &outBuf)
	err := io.MultiWriter(os.Stderr, &errBuf)

	cmd.Start()

	go func() {
		_, errOut = io.Copy(out, outIn)
	}()

	go func() {
		_, errErr = io.Copy(err, errIn)
	}()

	cmd.Wait()

	if errOut != nil || errErr != nil {
		log.Fatal("failed to capture stdout or stderr")
	}

	outStr := string(outBuf.Bytes())
	errStr := string(errBuf.Bytes())
	fmt.Println(outStr)
	fmt.Println(errStr)
}

func fileGlob(path string) []string {
	var list []string
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		re := regexp.MustCompile(`\.(go|h|c|s)$`)
		if re.MatchString(path) {
			list = append(list, path)
		}
		return nil
	})

	return list
}

func runtimeGoVersion() string {
	// go version go1.11.5 linux/amd64
	re := regexp.MustCompile(`go(\d[^ ]+)`)
	cmd, _ := exec.Command("/usr/bin/go", "version").Output()
	return re.FindStringSubmatch(string(cmd))[1]
}

func goVersionGreaterThan(v1, v2 string) bool {
	versions := []int64{}
	for _, v := range []string{v1, v2} {
		s := ""
		for _, i := range strings.Split(v, ".") {
			idx := 4 - len(i)
			s += strings.Repeat("0", idx) + i
		}
		num, _ := strconv.ParseInt(s, 10, 64)
		versions = append(versions, num)
	}
	return versions[0] > versions[1]
}

func buildRoot() string {
	storeArg("buildroot", os.Getenv("RPM_BUILD_ROOT"))
	return os.Getenv("RPM_BUILD_ROOT")
}

func buildDir() string {
	return os.Getenv("RPM_BUILD_DIR")
}

func goAbi() string {
	re := regexp.MustCompile(`go(\d+\.\d+).*`) // go1.8.3
	return re.FindStringSubmatch(runtime.Version())[1]
}

func libDir() string {
	out, e := exec.Command("rpm", "--eval", "%_libdir").Output()
	if e != nil {
		log.Fatal("Failed to call 'rpm --eval %_libdir', is rpm installed?")
	}
	return string(out)
}

func contribDir() string {
	return libDir() + "/go/" + goAbi() + "/contrib/pkg/linux_" + runtime.GOARCH
}

func toolDir() string {
	return "/usr/share/go/" + goAbi() + "/pkg/tool/linux_" + runtime.GOARCH
}

func contribSrcDir() string {
	return "/usr/share/go/" + goAbi() + "/contrib/src"
}

func buildPath() string {
	return buildDir() + "/go"
}

func buildContrib() string {
	return buildDir() + "/contrib"
}

func buildSrc() string {
	return buildContrib() + "/src"
}

func buildBin() string {
	return buildPath() + "/bin"
}

func destPath() string {
	return buildPath() + "/src/" + loadArg("importpath")
}

func arch() {
	fmt.Println(runtime.GOARCH)
}

func prep() {
	dirs := []string{destPath(), buildSrc(),
		buildRoot() + contribDir(),
		buildRoot() + contribSrcDir(),
		buildRoot() + toolDir(),
		buildRoot() + "/usr/bin"}
	// make dirs
	for _, d := range dirs {
		log.Println("Creating " + d)
		if _, e := os.Stat(d); e == nil {
			os.Remove(d)
		}
		os.MkdirAll(d, 0755)
	}

	// copy files
	currentDir, _ := os.Getwd()

	copyDir(currentDir, destPath())
	copyDir(contribSrcDir(), buildSrc())
}

func build() {
	buildFlags := []string{"-v", "-p", "4", "-x", "-buildmode=pie"}
	// Add s flag if go is older than 1.10.
	// s flag is an openSUSE flag to fix
	// https://bugzilla.suse.com/show_bug.cgi?id=776058
	// This flag is added with a patch in the openSUSE package, thus it only
	// exists in openSUSE go packages, and only on versions < 1.10.
	// In go >= 1.10, this bug is fixed upstream and the patch that was adding the
	// s flag has been removed from the openSUSE packages.
	if !goVersionGreaterThan(runtimeGoVersion(), "1.10.0") {
		buildFlags = append(buildFlags, "-s")
	}

	var extra []string
	var modifiers []string
	importpath := loadArg("importpath")
	extraFlags := loadArg("extraflags")
	modifier := loadArg("modifier")

	if strings.Contains(modifier, " ") {
		modifiers = strings.Split(modifier, " ")
	} else {
		modifiers = []string{modifier}
	}

	if len(extraFlags) > 0 {
		extra = strings.Split(extraFlags, " ")
	}

	args := append(buildFlags, extra...)

	for _, modifier := range modifiers {
		path := ""
		if modifier == "..." || modifier == "/..." {
			path = importpath + modifier
		} else {
			path = importpath + "/" + modifier
		}

		goBuild("install", args, path)
	}
}

func install() {
	binaries, _ := filepath.Glob(buildBin() + "/*")

	if len(binaries) > 0 {
		for _, bin := range binaries {
			fmt.Println("Copying " + bin)
			copyFile(bin, buildRoot()+"/usr/bin")
		}
	}
}

func source() {
	files := fileGlob(buildPath() + "/src")
	re := regexp.MustCompile(buildPath() + "/src" + `(.*)$`)
	for _, f := range files {
		dest := buildRoot() + contribSrcDir() + re.FindStringSubmatch(f)[1]

		if _, e := os.Stat(filepath.Dir(dest)); e != nil {
			os.MkdirAll(filepath.Dir(dest), 0755)
		}

		fmt.Println(f)
		fmt.Println(dest)
		copyFile(f, dest)
	}
}

func test() {
	var extra []string
	var modifiers []string

	importpath := loadArg("importpath")
	extraFlags := loadArg("extraflags")
	modifier := loadArg("modifier")

	if strings.Contains(modifier, " ") {
		modifiers = strings.Split(modifier, " ")
	} else {
		modifiers = []string{modifier}
	}

	if len(extraFlags) > 0 {
		extra = strings.Split(extraFlags, " ")
	}

	args := append(extra, "-x")

	for _, modifier := range modifiers {
		var path string
		if modifier == "..." || modifier == "/..." {
			path = importpath + modifier
		} else {
			path = importpath + "/" + modifier
		}

		goBuild("test", args, path)
	}
}

func filelist() {
	// list everything under buildroot
	var list []string
	filepath.Walk(buildRoot(), func(path string, info os.FileInfo, err error) error {
		re := regexp.MustCompile(regexp.QuoteMeta(buildRoot()) + `(.*)$`)
		fmt.Println(re.FindStringSubmatch(path))
		if len(re.FindStringSubmatch(path)) > 1 {
			if info.IsDir() {
				list = append(list, "%dir "+re.FindStringSubmatch(path)[1])
			} else {
				list = append(list, re.FindStringSubmatch(path)[1])
			}
		}
		return nil
	})

	fd, _ := os.Create("file.lst")
	defer fd.Close()

	for _, entry := range list {
		fd.WriteString(entry + "\n")
	}
}

func godoc() {
	fmt.Println("We should generate proper godocs!")
}

func main() {
	opts := os.Args
	size := len(opts)
	option := Option{}
	option.Load()
	action := ""

	supportedActions := map[string]func(){"arch": arch,
		"prep":     prep,
		"build":    build,
		"install":  install,
		"source":   source,
		"test":     test,
		"filelist": filelist,
		"godoc":    godoc}

	if size == 1 {
		// print help
		fmt.Println("Please specify a valid metho: arch, prep, build, install, source, test, filelist, godoc")
	}

	if size == 2 {
		action = opts[1]
	}

	if size > 2 {
		action = opts[1]
		if action == "prep" {
			option.Importpath = opts[2]
		}
		if action == "test" || action == "build" {
			option.Fill(opts[2:])
		}
	}

	if _, ok := supportedActions[action]; !ok {
		log.Fatalf("%s is not a supported action.", action)
	}

	option.Save()

	supportedActions[action]()
}
