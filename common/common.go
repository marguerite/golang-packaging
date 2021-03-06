package common

import (
	"fmt"
	"github.com/marguerite/golang-packaging/option"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
)

func goAbi() string {
	re := regexp.MustCompile(`go(\d+\.\d+).*`)
	return re.FindStringSubmatch(runtime.Version())[1]
}

// GoList run `go list` with env and options
func GoList(options []string, opt option.Option) string {
	flags := append([]string{"list"}, options...)

	cmd := exec.Command("/usr/bin/go", flags...)
	env := append(os.Environ(), "GOPATH="+goPath(opt))
	env = append(env, "GO15VENDOREXPERIMENT=1")
	cmd.Env = env

	out, _ := cmd.CombinedOutput()

	return string(out)
}

func goPath(opt option.Option) string {
	return opt.BuildRoot + "/usr/share/go/" + goAbi() + "/contrib" + ":" + opt.BuildPath
}

// ReadStdin read from stdin to avoid a broken pipe
func ReadStdin(args []string) {}

func libDir() string {
	out, e := exec.Command("rpm", "--eval", "%_libdir").Output()
	if e != nil {
		log.Fatal("Failed to call 'rpm --eval %_libdir', is rpm installed?")
	}
	return string(out)
}

// ContribDir go contrib dir
func ContribDir() string {
	return libDir() + "/go/" + goAbi() + "/contrib/pkg/linux_" + runtime.GOARCH
}

// ToolDir go tool dir
func ToolDir() string {
	return "/usr/share/go/" + goAbi() + "/pkg/tool/linux_" + runtime.GOARCH
}

// ContribSrcDir go contrib src dir
func ContribSrcDir() string {
	return "/usr/share/go/" + goAbi() + "/contrib/src"
}

// CopyDir copy files in a directory
func CopyDir(src, dst string) {
	fmt.Printf("Copying all files and directories from %s to %s...", src, dst)

	if _, e := os.Stat(dst); e != nil {
		fmt.Printf("%s doesn't exist, making...", dst)
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
				CopyFile(path, filepath.Join(dst, file))
			}
			if info.Mode().IsDir() {
				CopyDir(path, filepath.Join(dst, file))
			}
			if info.Mode()&os.ModeSymlink != 0 {
				symlink, e := os.Readlink(path)
				if e != nil {
					log.Fatalf("Could not follow %s's symlink", path)
				}

				// make absolute symlink
				if !filepath.IsAbs(symlink) {
					fmt.Printf("Non-absolute symlink found: %s", symlink)
					fmt.Println("Converting to absolute path")
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

				CopyFile(symlink, filepath.Join(dst, file))
			}
			return nil
		})
	}
}

// CopyFile copy a file
func CopyFile(src, dst string) {
	sd, e := os.Open(src)
	if e != nil {
		log.Fatalf("Failed to open %s file descriptor", src)
	}
	defer sd.Close()

	if f, e := os.Stat(dst); !os.IsNotExist(e) && f.Mode().IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	}

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
