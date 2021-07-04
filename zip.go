package main

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"
)

func createZip(handled *map[string]*hash) {
	for k, h := range *handled {
		v := h.version
		p := filepath.Join(GOMOD(), "cache/download", k, "@v", v)
		z := p + ".zip"
		f, _ := os.Create(z)
		zw := zip.NewWriter(f)

		i := strings.Index(k, "/")
		n := 1
		addFiles(zw, filepath.Join(GOMOD(), k[:i]), k[:i], k+"@"+v, n)
		zw.Close()
		f.Close()
		h1 := ziphash(z)
		fmt.Printf("\t%s dirhash: %s\n", z, h1)
		(*handled)[k].zip = h1

		f1, _ := os.Create(p + ".ziphash")
		f1.WriteString(h1)
		f1.Close()
	}
}

func addFiles(zw *zip.Writer, base, zbase, key string, n int) {
	arr := strings.Split(key, "/")
	f, _ := os.Open(base)
	defer f.Close()
	files, _ := f.Readdir(-1)

	for _, v := range files {
		if n < len(arr) && v.Name() != arr[n] {
			continue
		}
		if v.IsDir() {
			nb := filepath.Join(base, v.Name())
			addFiles(zw, nb, filepath.Join(zbase, v.Name()), key, n+1)
			continue
		}

		data, _ := ioutil.ReadFile(filepath.Join(base, v.Name()))
		f1, _ := zw.Create(filepath.Join(zbase, v.Name()))
		f1.Write(data)
	}
}

func ziphash(z string) string {
	h, _ := dirhash.HashZip(z, dirhash.DefaultHash)
	return h
}