package main

import (
	"bytes"
	"fmt"
	"github.com/marguerite/golang-packaging/common"
	"github.com/marguerite/golang-packaging/option"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

func goBuild(command string, options []string, path string, opt option.Option) {
	flags := append(append([]string{command}, options...), []string{path}...)

	var outBuf, errBuf bytes.Buffer
	var errOut, errErr error
	var wg sync.WaitGroup

	cmd := exec.Command("/usr/bin/go", flags...)
	env := append(os.Environ(), "GOPATH="+opt.BuildPath+":"+opt.BuildContrib)
	env = append(env, "GOBIN="+opt.BuildBin)
	cmd.Env = env

	log.Printf("Command: GOPATH=%s GOBIN=%s /usr/bin/go %s\n", opt.BuildPath+":"+opt.BuildContrib, opt.BuildBin, strings.Join(flags, " "))

	outIn, _ := cmd.StdoutPipe()
	errIn, _ := cmd.StderrPipe()

	out := io.MultiWriter(os.Stdout, &outBuf)
	err := io.MultiWriter(os.Stderr, &errBuf)

	cmd.Start()
	wg.Add(1)

	go func() {
		_, errOut = io.Copy(out, outIn)
		wg.Done()
	}()

	_, errErr = io.Copy(err, errIn)
	wg.Wait()

	cmd.Wait()

	if errOut != nil {
		log.Fatalf("Failed to capture stdout %s", errOut)
	}

	if errErr != nil {
		log.Fatalf("Failed to capture stderr %s", errErr)
	}

	fmt.Println(string(outBuf.Bytes()))
	fmt.Println(string(errBuf.Bytes()))
}

func sourceFileFilter(path string) []string {
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

func arch() {
	fmt.Println(runtime.GOARCH)
}

func prep(opt option.Option) {
	dirs := []string{opt.DestPath, opt.BuildSrc,
		filepath.Join(opt.BuildRoot, common.ContribDir()),
		filepath.Join(opt.BuildRoot, common.ContribSrcDir()),
		filepath.Join(opt.BuildRoot, common.ToolDir()),
		filepath.Join(opt.BuildRoot, "/usr/bin")}
	// make dirs
	for _, d := range dirs {
		fmt.Println("Creating " + d)
		os.MkdirAll(d, 0755)
	}

	// copy files
	currentDir, _ := os.Getwd()

	common.CopyDir(currentDir, opt.DestPath)
	common.CopyDir(common.ContribSrcDir(), opt.BuildSrc)
}

func build(opt option.Option) {
	buildFlags := []string{"-v", "-p", "4", "-x"}
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

	buildFlags = append(buildFlags, opt.BuildMode)

	args := append(buildFlags, opt.ExtraFlags...)

	if len(opt.Modifiers) > 0 {
		for _, modifier := range opt.Modifiers {
			path := ""
			if modifier == "..." {
				path = opt.ImportPath + modifier
			} else {
				path = opt.ImportPath + "/" + modifier
			}

			goBuild("install", args, path, opt)
		}
	} else {
		goBuild("install", args, opt.ImportPath+"...", opt)
	}
}

func install(opt option.Option) {
	binaries, _ := filepath.Glob(opt.BuildBin + "/*")

	if len(binaries) > 0 {
		for _, bin := range binaries {
			dest := filepath.Join(opt.BuildRoot, "/usr/bin")
			fmt.Printf("Copying %s to %s\n", bin, dest)
			common.CopyFile(bin, dest)
		}
	}
}

func source(opt option.Option) {
	re := regexp.MustCompile(opt.BuildPath + "/src" + `(.*)$`)
	files := sourceFileFilter(opt.BuildPath + "/src")
	for _, f := range files {
		dest := opt.BuildRoot + common.ContribSrcDir() + re.FindStringSubmatch(f)[1]

		if _, e := os.Stat(filepath.Dir(dest)); os.IsNotExist(e) {
			os.MkdirAll(filepath.Dir(dest), 0755)
		}
		fmt.Printf("Copying %s to %s\n", f, dest)
		common.CopyFile(f, dest)
	}
}

func test(opt option.Option) {
	args := append([]string{"-x"}, opt.ExtraFlags...)

	if len(opt.Modifiers) > 0 {
		for _, modifier := range opt.Modifiers {
			var path string
			if modifier == "..." {
				path = opt.ImportPath + modifier
			} else {
				path = opt.ImportPath + "/" + modifier
			}
			goBuild("test", args, path, opt)
		}
	} else {
		goBuild("test", args, opt.ImportPath+"...", opt)
	}
}

func filelist(opt option.Option) {
	// list everything under buildroot
	var list []string
	filepath.Walk(opt.BuildRoot, func(path string, info os.FileInfo, err error) error {
		re := regexp.MustCompile(regexp.QuoteMeta(opt.BuildRoot) + `(.*)$`)
		if len(re.FindStringSubmatch(path)) > 1 {
			p := re.FindStringSubmatch(path)[1]
			if info.IsDir() {
				fmt.Printf("%%dir %s\n", p)
				list = append(list, "%dir "+p)
			} else {
				fmt.Println(p)
				list = append(list, p)
			}
		}
		return nil
	})

	cwd, _ := os.Getwd()
	path := filepath.Join(cwd, "file.list")

	fd, e := os.Create(path)
	if e != nil {
		log.Fatalf("Failed to create file.list in %s", cwd)
	}
	defer fd.Close()

	for _, entry := range list {
		fd.WriteString(entry + "\n")
	}
}

func godoc() {
	fmt.Println("We should generate proper godocs!")
}

func main() {
	args := os.Args

	if len(args) == 1 {
		// print help
		fmt.Println("Please specify a valid method: arch, prep, build, install, source, test, filelist, godoc")
		os.Exit(0)
	}

	//action := args[1]

	opt := option.Option{}
	opt.Load()
	if len(args) >= 2 {
		opt.Parse(args[2:])
	}

	if len(opt.BuildRoot) == 0 {
		opt.BuildRoot = os.Getenv("RPM_BUILD_ROOT")
	}

	if len(opt.BuildDir) == 0 {
		opt.BuildDir = os.Getenv("RPM_BUILD_DIR")
	}

	opt.BuildPath = filepath.Join(opt.BuildDir, "go")
	opt.BuildContrib = filepath.Join(opt.BuildDir, "contrib")
	opt.BuildSrc = filepath.Join(opt.BuildContrib, "src")
	opt.BuildBin = filepath.Join(opt.BuildPath, "bin")
	opt.DestPath = filepath.Join(opt.BuildPath, "/src/"+opt.ImportPath)
	opt.Save()

	/*switch action {
	case "arch":
		arch()
	case "prep":
		prep(opt)
	case "build":
		build(opt)
	case "install":
		install(opt)
	case "source":
		source(opt)
	case "test":
		test(opt)
	case "filelist":
		filelist(opt)
	case "godoc":
		godoc()
	default:
		log.Fatalf("%s is not a supported action.", action)
	}*/
}
