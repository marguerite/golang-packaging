package main

import (
	"fmt"
	"github.com/marguerite/golang-packaging/common"
	"github.com/marguerite/golang-packaging/option"
	"os"
	"strings"
)

func main() {
	common.Skeleton(os.Args)
	opt := option.Option{}
	opt.Load()
	options := []string{"-f", "'{{.ImportPath}}'", opt.ImportPath + "/..."}
	for _, i := range strings.Split(common.Exec(options, opt), "\n") {
		if !strings.Contains(i, "/vendor/") && !strings.Contains(i, "matched no packages") && len(i) > 0 {
			fmt.Println("golang(" + i + ")")
		}
	}
}
