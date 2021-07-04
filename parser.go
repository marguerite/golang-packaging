package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"unicode"

	"golang.org/x/mod/module"
)

func parseImports(files []string, importpath string) map[string]struct{} {
	imports := make(map[string]struct{})
	ch := make(chan string)
	var wg sync.WaitGroup

	for _, v := range files {
		wg.Add(1)
		go func(file string) {
			f, err := os.Open(file)
			if err != nil {
				panic(err)
			}
			defer func() {
				f.Close()
				wg.Done()
			}()
			parseImport(f, importpath, ch)
		}(v)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for {
		val, ok := <-ch
		if !ok {
			break
		}
		if _, ok = imports[val]; !ok {
			imports[val] = struct{}{}
		}
	}

	return imports

}

func parseImport(r io.ReadSeeker, importpath string, ch chan string) {
	var found bool
	for {
		buf, err := readLine(r)
		if err == io.EOF {
			break
		}
		if bytes.HasPrefix(buf, []byte("import")) {
			if i := bytes.IndexByte(buf, '('); i > 0 {
				found = true
				continue
			}
			buf = buf[8 : len(buf)-2]
			if err := module.CheckPath(string(buf)); err == nil {
				if !bytes.HasPrefix(buf, []byte(importpath)) {
					ch <- string(buf)
				}
			}
		}

		if bytes.HasPrefix(buf, []byte{')'}) && found {
			found = false
			break
		}

		if found {
			// //"github.com/gavv/monotime"
			if bytes.Contains(buf, []byte("//")) {
				continue
			}
			buf = bytes.TrimFunc(buf, func(r rune) bool {
				return unicode.IsSpace(r) || unicode.IsPunct(r)
			})
			if i := bytes.Index(buf, []byte{' '}); i > 0 {
				buf = buf[i+2:]
			}

			if err := module.CheckPath(string(buf)); err == nil {
				if !bytes.HasPrefix(buf, []byte(importpath)) {
					ch <- string(buf)
				}
			}
		}
	}
}

type moduleline struct {
	explicit bool
	version  string
}

// parseModulesTxt parse the "modules.txt" in "vendor" directory
func parseModulesTxt(txt string) map[string]moduleline {
	f, err := os.Open(txt)
	if err != nil {
		fmt.Printf("can not find %s, please do not remove it when local compressing\n", txt)
		os.Exit(1)
	}
	defer f.Close()

	m := make(map[string]moduleline)

	for {
		b, err := readLine(f)
		if err == io.EOF {
			break
		}

		if b[0] == '#' && b[1] == ' ' {
			b = b[2:]
			i := bytes.Index(b, []byte{' '})
			b1, _ := readLine(f)
			if b1[1] == '#' {
				m[string(b[:i])] = moduleline{true, string(b[i+1 : len(b)-1])}
			} else {
				m[string(b[:i])] = moduleline{false, string(b[i+1 : len(b)-1])}
			}
		}
	}

	return m
}

func readLine(r io.ReadSeeker) ([]byte, error) {
	var buf []byte

	for {
		b := make([]byte, 20)
		n, err := r.Read(b)

		if n < 20 {
			b = b[:n]
		}

		if err == io.EOF {
			return buf, err
		}

		if i := bytes.IndexByte(b, '\n'); i >= 0 {
			if i == 19 {
				for _, v := range b {
					buf = append(buf, v)
				}
				return buf, nil
			}
			for _, v := range b[:i+1] {
				buf = append(buf, v)
			}
			r.Seek(int64(i-len(b)+1), 1)
			return buf, nil
		}
		for _, v := range b {
			buf = append(buf, v)
		}
	}
	return buf, nil
}
