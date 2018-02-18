package main

import (
  "bytes"
  "fmt"
  "io"
  "io/ioutil"
  "os"
  "os/exec"
  "regexp"
  "runtime"
  "strings"
)

func check(e error) {
  if e != nil { panic(e) }
}

func go_abi() string {
  v := runtime.Version()
  re := regexp.MustCompile(`go(\d+\.\d+).*`)
  return re.FindStringSubmatch(v)[1]
}

func checkout_arg(path string) string {
  for i, p := range []string{"buildroot", "importpath"} {
    if path == p { break }
    if i == 1 { panic("Only support: buildroot and importpath") }
  }

  if _, e := os.Stat("/tmp/" + path); e == nil {
    f, e := os.OpenFile("/tmp/" + path, os.O_RDWR, 0644)
    check(e)
    defer f.Close()

    s, e := ioutil.ReadAll(f)
    check(e)

    return string(s)
  } else {
    panic(e)
  }
}

func run_cmd(options []string) string {
  flags := append([]string{"list"}, options...)

  var outBuf bytes.Buffer
  var errOut error

  cmd := exec.Command("/usr/bin/go", flags...)
  env := append(os.Environ(), "GOPATH=" + gopath())
  env = append(env, "GO15VENDOREXPERIMENT=1")
  cmd.Env = env

  outIn, _ := cmd.StdoutPipe()

  out := io.MultiWriter(os.Stdout, &outBuf)

  e := cmd.Start()
  check(e)

  go func() {
    _, errOut = io.Copy(out, outIn)
  }()

  e = cmd.Wait()
  check(e)

  if errOut != nil {
    panic("failed to capture stdout\n")
  }

  return string(outBuf.Bytes())
}

func gopath() string {
  var paths []string
  var gopath string
  for _, path := range strings.Split(os.Getenv("GOPATH"), ":") {// /home/abuild/go:/usr/share/go/1.9/contrib
    re := regexp.MustCompile(`contrib`)
    if !re.MatchString(path) {
      paths = append(paths, path)
    }
  }

  for _, path := range paths {
    gopath += ":" + path
  }

  return checkout_arg("buildroot") + "/usr/share/go/" + go_abi() + "/contrib" + gopath
}

func skeleton(args []string) {
  // read from stdin to not cause a broken pipe
}

func main() {
  skeleton(os.Args)

  options := []string{"-f", "'{{.ImportPatH}}'", checkout_arg("importpath") + "/..."}
  str := run_cmd(options)
  items := strings.Split(str, "\n")

  re := regexp.MustCompile(`/vendor/`)

  for _, i := range items {
    if !re.MatchString(i) {
      fmt.Println("golang(" + i + ")")
    }
  }
}
