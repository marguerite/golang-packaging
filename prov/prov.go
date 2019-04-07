package main

import (
	"fmt"
	"github.com/marguerite/golang-packaging/common"
	"github.com/marguerite/golang-packaging/option"
	"os"
	"strings"
)

func main() {
	common.ReadStdin(os.Args)
	opt := option.Option{}
	opt.Load()
	options := []string{"-f", "'{{.ImportPath}}'", opt.ImportPath + "/..."}
	r := strings.NewReplacer("'", "", " ", "")
	for _, i := range strings.Split(common.GoList(options, opt), "\n") {
		if !strings.Contains(i, "/vendor/") && !strings.Contains(i, "matched no packages") {
			i = r.Replace(i)
			if len(i) > 0 {
				fmt.Println("golang(" + i + ")")
			}
		}
	}
}
