package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/marguerite/go-stdlib/dir"
	"github.com/marguerite/go-stdlib/fileutils"
	"golang.org/x/mod/sumdb/dirhash"
)

func fakeGoModBatch(handled *map[string]*hash) {
	for k, h := range *handled {
		v := h.version
		files, _ := dir.Glob("**/*.go", filepath.Join(GOMOD(), k+"@"+v))
		imports := parseImports(files, k)
		gomod := "module " + k + "\n\ngo " + Version() + "\n\n"
		if len(imports) > 0 {
			gomod += "require (\n"
			imports1 := make(map[string]string)
			for k1 := range imports {
				b := base(k1)
				if _, ok := imports1[b]; ok {
					continue
				}
				if _, ok := BUILDREQUIRES[b]; ok {
					continue
				}
				if val, ok := VENDORED[b]; ok {
					imports1[b] = val.version
					continue
				}
				fmt.Printf("uncovered import %s\n", k1)
				os.Exit(1)
			}

			for k1, v1 := range imports1 {
				gomod += "\t" + k1 + " " + v1 + "\n"
			}

			gomod += ")\n"
		}

		(*handled)[k].gomod = modhash(gomod)

		fmt.Printf("Faking %s\n", filepath.Join(GOMOD(), k+"@"+v, "go.mod"))
		f, err := os.Create(filepath.Join(GOMOD(), k+"@"+v, "go.mod"))
		if err != nil {
			panic(err)
		}
		d := filepath.Join(GOMOD(), "cache/download/", k, "@v")
		dir.MkdirP(d)
		fmt.Printf("Faking %s\n", filepath.Join(GOMOD(), "cache/download/", k, "@v", v+".mod"))
		f1, err := os.Create(filepath.Join(GOMOD(), "cache/download/", k, "@v", v+".mod"))
		if err != nil {
			panic(err)
		}
		f.WriteString(gomod)
		f1.WriteString(gomod)
		f.Close()
		f1.Close()
	}
}

func modhash(gomod string) string {
	h, _ := dirhash.Hash1([]string{"go.mod"}, func(string) (io.ReadCloser, error) {
		return ioutil.NopCloser(strings.NewReader(gomod)), nil
	})
	return h
}

func fakeGoMod(dst string) {
	gomod := "module " + IMPORTPATH + "\n\ngo " +
		Version() + "\n\nrequire (\n"
	for k, v := range VENDORED {
		if _, ok := BUILDREQUIRES[base(k)]; v.explicit && !ok {
			gomod += "\t" + k + " " + v.version + "\n"
		}
	}
	for k, v := range BUILDREQUIRES {
		gomod += "\t" + k + " " + v + "\n"
	}
	gomod += ")\n"
	fmt.Println(gomod)

	// rename old
	if _, err := os.Stat(dst); err == nil {
		err := os.Rename(dst, dst+".bak")
		if err != nil {
			panic(err)
		}
	}
	f, err := os.Create(dst)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	n, err := f.WriteString(gomod)
	if err != nil {
		panic(err)
	}
	if n != len(gomod) {
		fmt.Printf("partially written file %s\n", dst)
		os.Exit(1)
	}
}

func fakeGoSum(dst string, handled *map[string]*hash) {
	// rename old
	if _, err := os.Stat(dst); err == nil {
		err := os.Rename(dst, dst+".bak")
		if err != nil {
			panic(err)
		}

		f, _ := os.Open(dst + ".bak")
		f1, _ := os.Create(dst)

		for {
			buf, err := readLine(f)
			if err == io.EOF {
				break
			}
			i := bytes.Index(buf, []byte{' '})
			s := buf[:i]
			buf = buf[i+1:]
			f1.Write(s)
			f1.Write([]byte{' '})
			if _, ok := (*handled)[string(s)]; ok && bytes.HasPrefix(buf, []byte((*handled)[string(s)].version)) {
				i = bytes.Index(buf, []byte{' '})
				buf = buf[:i]
				f1.Write(buf)
				f1.Write([]byte{' '})
				if bytes.HasSuffix(buf, []byte("/go.mod")) {
					f1.WriteString((*handled)[string(s)].gomod)
				} else {
					f1.WriteString((*handled)[string(s)].zip)
				}
				f1.Write([]byte{'\n'})
			} else {
				f1.Write(buf)
			}
		}

		f.Close()
		f1.Close()
		return
	}

	f1, _ := os.Create(dst)
	for k, h := range *handled {
		f1.WriteString(k + " " + h.version + " " + h.zip + "\n")
		f1.WriteString(k + " " + h.version + "/go.mod " + h.gomod + "\n")
	}
	f1.Close()
}

func base(k string) string {
	if strings.Count(k, "/") > 2 {
		k1 := k
		var x int
		for i := 0; i < 3; i++ {
			j := strings.Index(k1, "/")
			k1 = k1[j+1:]
			x += j + 1
		}
		return k[:x-1]
	}
	return k
}

type hash struct {
	version string
	gomod   string
	zip     string
}

func fakeMods(imports map[string]struct{}, handled *map[string]*hash) {
	for k := range imports {
		b := base(k)
		if val, ok := BUILDREQUIRES[b]; ok {
			files, _ := dir.Glob(filepath.Join("/usr/share/go/*/contrib/src/", b))
			copyDir(k, files[0], filepath.Join(GOMOD(), b+"@"+val, strings.TrimPrefix(k, b)), handled)
			(*handled)[b] = &hash{val, "", ""}
			continue
		}
		if val, ok := VENDORED[b]; ok {
			copyDir(k, filepath.Join(GOSRC(), IMPORTPATH, "vendor", b, strings.TrimPrefix(k, b)), filepath.Join(GOMOD(), b+"@"+val.version, strings.TrimPrefix(k, b)), handled)
			(*handled)[b] = &hash{val.version, "", ""}
			continue
		}
		fmt.Printf("Cannot find module %s in either BuildRequires or vendor directory\n", k)
		os.Exit(1)
	}
}

func copyDir(importpath, src, dst string, handled *map[string]*hash) {
	fmt.Printf("Copying files from %s to %s\n", src, dst)
	err := dir.MkdirP(dst)
	if err != nil && !os.IsExist(err) {
		return
	}
	files, _ := dir.Glob("*.go", src, "*_test.go")
	files = buildable(files)
	for _, v := range files {
		if _, err := os.Stat(filepath.Join(dst, filepath.Base(v))); os.IsNotExist(err) {
			fmt.Printf("Copying %s\n", v)
			fileutils.Copy(v, dst)
		} else {
			fmt.Printf("Skipping existing %s\n", v)
		}
	}
	imports := parseImports(files, importpath)
	if len(imports) > 0 {
		fakeMods(imports, handled)
	}
}

func buildable(files []string) []string {
	var files1 []string
	for _, v := range files {
		f, _ := os.Open(v)
		b := make([]byte, 10)
		f.Read(b)
		if !bytes.Equal(b, []byte("// +build ")) {
			f.Close()
			files1 = append(files1, v)
			continue
		}
		var n int64
		n = 10
		for {
			b1 := make([]byte, 10)
			f.ReadAt(b1, n)
			if i := bytes.IndexByte(b1, ' '); i >= 0 {
				if i == 0 {
					n++
					continue
				}
				b1 = b1[:i]
				if bytes.Equal(b1, []byte("linux")) || bytes.Equal(b1, []byte("unix")) {
					files1 = append(files1, v)
					break
				}
				n += int64(i)
				continue
			}
			if i := bytes.IndexByte(b1, '\n'); i >= 0 {
				if i == 0 {
					break
				}
				b1 = b1[:i]
				if bytes.Equal(b1, []byte("linux")) || bytes.Equal(b1, []byte("unix")) {
					files1 = append(files1, v)
				}
				break
			}
		}
		f.Close()
	}
	return files1
}
