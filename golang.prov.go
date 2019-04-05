package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func goAbi() string {
	re := regexp.MustCompile(`go(\d+\.\d+).*`)
	return re.FindStringSubmatch(runtime.Version())[1]
}

func loadArg(arg string) string {
	supportedArgs := map[string]struct{}{"buildroot": {}, "importpath": {}}

	// Error Messages have to one word, or RPM dependency generator will split it
	if _, ok := supportedArgs[arg]; !ok {
		panic("OnlySupportBuildrootOrImportpath")
	}

	argFile := filepath.Join("/tmp", arg)

	if _, e := os.Stat(argFile); os.IsNotExist(e) {
		fmt.Printf("FileNotExist%s", argFile)
		panic(e)
	}

	fd, e := os.OpenFile(argFile, os.O_RDWR, 0644)
	if e != nil {
		fmt.Printf("PermissionErrorFailedToOpen%s", argFile)
		panic(e)
	}
	defer fd.Close()

	b, e := ioutil.ReadAll(fd)
	if e != nil {
		fmt.Printf("CouldNotReadContentOf%s", argFile)
		panic(e)
	}

	return string(b)
}

func run(options []string) string {
	flags := append([]string{"list"}, options...)

	cmd := exec.Command("/usr/bin/go", flags...)
	env := append(os.Environ(), "GOPATH="+goPath())
	env = append(env, "GO15VENDOREXPERIMENT=1")
	cmd.Env = env

	out, e := cmd.CombinedOutput()
	if e != nil {
		panic("NotInstalled/usr/bin/go")
	}

	return string(out)
}

func goPath() string {
	paths := []string{}
	for _, path := range strings.Split(os.Getenv("GOPATH"), ":") { // /home/abuild/go:/usr/share/go/1.9/contrib
		re := regexp.MustCompile(`contrib`)
		if !re.MatchString(path) {
			paths = append(paths, path)
		}
	}

	return loadArg("buildroot") + "/usr/share/go/" + goAbi() + "/contrib" + strings.Join(paths, ":")
}

func skeleton(args []string) {
	// read from stdin to not cause a broken pipe
}

func main() {
	skeleton(os.Args)

	options := []string{"-f", "'{{.ImportPath}}'", loadArg("importpath") + "/..."}
	str := run(options)

	items := strings.Split(str, "\n")
	re := regexp.MustCompile(`/vendor/`)

	for _, i := range items {
		if !re.MatchString(i) && i != "" {
			fmt.Println("golang(" + i + ")")
		}
	}
}
