package main

import (
	"fmt"
	"github.com/marguerite/golang-packaging/common"
	"github.com/marguerite/golang-packaging/option"
	"os"
	"regexp"
	"strings"
)

func isStdLib(lib string, opt option.Option) bool {
	options := []string{"-f", "'{{if not .Standard}}{{.ImportPath}}{{end}}'", lib}
	if len(common.Exec(options, opt)) > 0 {
		return false
	}
	return true
}

func main() {
	common.Skeleton(os.Args)
	opt := option.Option{}
	opt.Load()
	options := []string{"-f", `'{{range $deps := .Deps}}{{printf "%s\n" $deps}}{{end}}'`, opt.ImportPath + "/..."}
	re := regexp.MustCompile(regexp.QuoteMeta(opt.ImportPath))
	r := strings.NewReplacer("'", "", " ", "")
	for _, i := range strings.Split(common.Exec(options, opt), "\n") {
		if !re.MatchString(i) && !strings.Contains(i, "matched no packages") {
			i = r.Replace(i)
			if !isStdLib(i, opt) && len(i) > 0 {
				fmt.Println("golang(" + i + ")")
			}
		}
	}
}
