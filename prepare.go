package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marguerite/go-stdlib/dir"
)

func prepareBuildEnvironment(args []string, module bool) {
	fmt.Println("Preparating build...")
	importpath, err := parseImportPath(args)
	if err != nil {
		panic(err)
	}
	storeImportPath(importpath)

	fmt.Printf("Parsed importpath '%s'\n", importpath)

	fmt.Println("Creating common direcotries...")

	arr := strings.Split(importpath, "/")
	err = dir.MkdirP(filepath.Join(GOSRC(), strings.Join(arr[:len(arr)-1], "/")))
	if err != nil {
		panic(err)
	}

	builddir := filepath.Join(GOSRC(), importpath)

	err = os.Symlink(RPM_BUILD_DIR(), builddir)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s created\n", builddir)

	err = dir.MkdirP(GOBIN())
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s created\n", GOBIN())

	if module {
		fmt.Println("Moduleaware mode detected, linking vendor to mod")

		if _, err := os.Stat(filepath.Join(RPM_BUILD_DIR(), "vendor")); os.IsNotExist(err) {
			fmt.Println("Moduleaware mode requires a vendor directory as reference to fake the go modules since we can't download go modules from an offline build virtual machine.")
			fmt.Printf("Please run `go mod init %s && go mod tidy && go mod vendor` locally in the unpacked source, or run `go mod download && go mod vendor` instead if the unpacked source contains a go.mod\n", importpath)
			fmt.Println("Then upload the vendor directory tarball as Source1, use `-a1` flag in the `setup` line to automatically unpack it.")
			os.Exit(1)
		}

		// below is the standard golang approach, of course we can just use "-mod=vendor" in `go build` command
		// but as the author I want to leave the possibility to users. so if users explicitly use -mod=vendor,
		// they can skip the BuildRequires fake step. if we take `-mod=vendor`, of course we have to directly
		// do the BuildRequires fake inside the vendor directory. then users have no way to go.

		// first step: if to replace any module with the one in BuildRequires
		// second step: fake a go.mod for the modules
		// third step: create a zip for the module directory and calculate the sha1 hash
		// actually copy the zip and module directory
		fmt.Println("These vendored modules found:")
		for k, v := range VENDORED {
			fmt.Printf("\t%s %s\n", k, v.version)
		}

		files, err := dir.Glob("**/*.go", RPM_BUILD_DIR(), "vendor/**/*.go")

		if err != nil {
			panic(err)
		}

		imports := parseImports(files, importpath)

		fmt.Println("These foreign imports found:")
		for k := range imports {
			fmt.Printf("\t%s\n", k)
		}

		fmt.Println("These BuildRequires found:")

		for k, v := range BUILDREQUIRES {
			fmt.Printf("\t%s %s\n", k, v)
		}

		fmt.Println("As we are RPM based distribution, BuildRequires always take priority over vendor directory, BuildRequires version will be used unless explicitly build with -mod=vendor")

		fmt.Println("Faking main go.mod with BuildRequires and vendor directory")

		fakeGoMod(filepath.Join(builddir, "go.mod"), importpath)

		fmt.Println("Faking modules with BuildRequires and vendor directory")

		//defer os.RemoveAll(tmp)
		handled := make(map[string]*hash)
		fakeMods(importpath, imports, &handled)

		fmt.Println("Faking dependencies' go.mod")
		fakeGoModBatch(&handled)
		fmt.Println("Creating zip file and zip hash")
		createZip(&handled)
		fakeGoSum(filepath.Join(builddir, "go.sum"), &handled)
	} else {
		fmt.Println("Go Path mode detected, linking vendor to src")

		/*for _, v := range matches {
			f, _ := os.Stat(v)
			if f.IsDir() {

			}
		}*/
	}
}
