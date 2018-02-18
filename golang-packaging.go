package main

import (
  "bytes"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "os"
  "os/exec"
  "path/filepath"
  "regexp"
  "runtime"
  "strings"
)

func check(e error) {
  if e != nil { panic(e) }
}

func copy_dir(s, d string) {
  if _, e := os.Stat(s); e != nil {
    return
  }

  if _, e := os.Stat(d); e != nil {
    os.MkdirAll(d, 0755)
  }

  sd, e := os.Open(s);
  check(e)
  defer sd.Close()

  files, e := sd.Readdirnames(-1)
  check(e)

  for _, f := range files {
    m, e := os.Stat(filepath.Join(s, f))
    check(e)
    if m.Mode().IsDir() {
      copy_dir(filepath.Join(s, f), filepath.Join(d, f))
    } else {
      copy_file(filepath.Join(s, f), filepath.Join(d, f))
    }
  }
}

func copy_file(s, d string) {
  fmt.Println("Copying " + s + " to " + d)
  sf, e := os.Open(s)
  check(e)
  defer sf.Close()

  f, e := os.Create(d)
  check(e)
  defer f.Close()

  _, e = io.Copy(f, sf)
  check(e)

  if m, e := os.Stat(s); e != nil {
    e = os.Chmod(d, m.Mode())
    check(e)
  }
}

func parse_args(args []string) (string,string,string) {
  index := 0
  var importpath, modifier, extraflags string
  // loop the args to find the first with "-"
  for i, arg := range args {
    re := regexp.MustCompile("-.*")
    ok := re.MatchString(arg)
    if ok {
      index = i
      break
    }
  }

  // form the extraflags
  if index > 0 {
    for _, arg := range args[index:] {
      extraflags += arg + " "
    }
  }

  var other []string
  if index > 0 {
    other = args[:index]
  } else {
    other = args
  }

  for i, arg := range other {
    if i == 0 {
      // split importpath from modifiers
      re := regexp.MustCompile(`(.*\/.*\/\w+)(.*)`)
      m := re.FindStringSubmatch(arg)
      if len(m) > 0 {
        importpath = m[1]
        modifier = m[2]
      } else {
        modifier = arg
      }
    } else {
      // "foo ..." equals to "foo..."
      re := regexp.MustCompile(`.*\w$`)
      if re.MatchString(modifier) && arg == "..." {
        modifier += arg
        // "foo bar ... baz" should be kept while bar in "foo ... bar" should be ignored
        if i == 1 {
          break
        }
      } else {
        modifier += " " + arg
      }
    }
  }

  return importpath, modifier, extraflags
}

func checkout_arg(arg string) (string,error) {
  args := []string{"importpath", "modifier", "extraflags"}

  for i, a := range args {
    if a == arg { break }

    if i == len(args) - 1 {
      panic("Only support: importpath, modifier, extraflags")
    }
  }

  path := "/tmp/" + arg
  if _, e := os.Stat(path); e != nil {
    return "", e
  }

  f, e := os.OpenFile(path, os.O_RDWR, 0644)
  check(e)
  defer f.Close()

  s, e := ioutil.ReadAll(f)
  check(e)

  return string(s), nil
}

func store(path string, content string) {
  path = "/tmp/" + path
  // create it if doesn't exist
  if _, e := os.Stat(path); e != nil {
    f, e := os.Create(path)
    check(e)
    defer f.Close()
    f.WriteString(content)
  }
}

func go_build(command string, options []string, path string) {
  flags := append(append([]string{command}, options...), []string{path}...)

  var outBuf, errBuf bytes.Buffer
  var errOut, errErr error

  cmd := exec.Command("/usr/bin/go", flags...)
  env := append(os.Environ(), "GOPATH=" + buildpath() + ":" + buildcontrib())
  env = append(env, "GOBIN=" + buildbin())
  cmd.Env = env

  outIn, _ := cmd.StdoutPipe()
  errIn, _ := cmd.StderrPipe()

  out := io.MultiWriter(os.Stdout, &outBuf)
  err := io.MultiWriter(os.Stderr, &errBuf)

  e := cmd.Start()
  check(e)

  go func() {
    _, errOut = io.Copy(out, outIn)
  }()

  go func() {
    _, errErr = io.Copy(err, errIn)
  }()

  e = cmd.Wait()
  check(e)

  if errOut != nil || errErr != nil {
    panic("failed to capture stdout or stderr\n")
  }

  outStr := string(outBuf.Bytes())
  errStr := string(errBuf.Bytes())
  fmt.Println(outStr)
  fmt.Println(errStr)
}

func file_glob(path string) []string {
  var list []string
  e := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
         re := regexp.MustCompile(`\.(go|h|s)$`)
         if re.MatchString(path) {
           list = append(list,path)
         }
         return nil
       })
  check(e)

  return list
}

func buildroot() string {
  store("buildroot", os.Getenv("RPM_BUILD_ROOT"))
  return os.Getenv("RPM_BUILD_ROOT")
}

func builddir() string {
  return os.Getenv("RPM_BUILD_DIR")
}

func go_abi() string {
  v := runtime.Version() // go1.8.3
  re := regexp.MustCompile(`go(\d+\.\d+).*`)
  return re.FindStringSubmatch(v)[1]
}

func libdir() string {
  out, e := exec.Command("rpm", "--eval", "%_libdir").Output()
  check(e)
  return string(out)
}

func contribdir() string {
  return libdir() + "/go/" + go_abi() + "/contrib/pkg/linux_" + runtime.GOARCH
}

func tooldir() string {
  return "/usr/share/go/" + go_abi() + "/pkg/tool/linux_" + runtime.GOARCH
}

func contribsrcdir() string {
  return "/usr/share/go/" + go_abi() + "/contrib/src"
}

func buildpath() string {
  return builddir() + "/go"
}

func buildcontrib() string {
  return builddir() + "/contrib"
}

func buildsrc() string {
  return buildcontrib() + "/src"
}

func buildbin() string {
  return buildpath() + "/bin"
}

func destpath() string {
  if importpath, e := checkout_arg("importpath"); e == nil {
    return buildpath() + "/src/" + importpath
  }
  panic("importpath may not found")
}

func arch() {
  fmt.Println(runtime.GOARCH)
}

func prep() {
  dirs := []string{destpath(), buildsrc(),
                   buildroot() + contribdir(),
                   buildroot() + contribsrcdir(),
                   buildroot() + tooldir(),
                   buildroot() + "/usr/bin"}
  // make dirs
  for _, d := range dirs {
    log.Println("Creating " + d)
    if _, e := os.Stat(d); e == nil {
      os.Remove(d)
    }
    e := os.MkdirAll(d, 0755)
    check(e)
  }

  // copy files
  current_dir, e := os.Getwd()
  check(e)

  copy_dir(current_dir, destpath())
  copy_dir(contribsrcdir(), buildsrc())
}


func build() {
  buildflags := []string{"-s", "-v", "-p", "4", "-x", "-buildmode=pie"}
  var extra []string
  var modifiers []string

  importpath, e := checkout_arg("importpath")
  s, e := checkout_arg("extraflags")
  m, e := checkout_arg("modifier")
  check(e)

  re := regexp.MustCompile(`\s`)
  if re.MatchString(m) {
    modifiers = strings.Split(m, "")
  } else {
    modifiers = []string{m}
  }

  if s != "" {
    extra = strings.Split(s, "")
  }

  args := append(buildflags, extra...)

  for _, modifier := range modifiers {
    var path string
    if modifier == "..." || modifier == "/..." {
      path = importpath + modifier
    } else {
      path = importpath + "/" + modifier
    }

    go_build("install", args, path)
  }
}

func install() {
  binaries, e := filepath.Glob(buildbin() + "/*")
  check(e)

  if len(binaries) > 0 {
    for _, bin := range binaries {
      fmt.Println("Copying " + bin)
      copy_file(bin, buildroot() + "/usr/bin")
    }
  }
}

func source() {
  files := file_glob(buildpath() + "/src")
  re := regexp.MustCompile(buildpath() + "/src" + `(.*)$`)
  for _, f := range files {
    dest := buildroot() + contribsrcdir() + re.FindStringSubmatch(f)[1]

    if _, e := os.Stat(filepath.Dir(dest)); e != nil {
      os.MkdirAll(filepath.Dir(dest), 0755)
    }

    fmt.Println(f)
    fmt.Println(dest)
    copy_file(f, dest)
  }
}

func test() {
  var extra []string
  var modifiers []string

  importpath, e := checkout_arg("importpath")
  s, e := checkout_arg("extraflags")
  m, e := checkout_arg("modifier")
  check(e)

  re := regexp.MustCompile(`\s`)
  if re.MatchString(m) {
    modifiers = strings.Split(m, "")
  } else {
    modifiers = []string{m}
  }

  if s != "" {
    extra = strings.Split(s, "")
  }

  args := append(extra, "-x")

  for _, modifier := range modifiers {
    var path string
    if modifier == "..." || modifier == "/..." {
      path = importpath + modifier
    } else {
      path = importpath + "/" + modifier
    }

    go_build("test", args, path)
  }
}


func filelist() {
  // list everything under buildroot
  var list []string
  e := filepath.Walk(buildroot(), func(path string, info os.FileInfo, err error) error {
         re := regexp.MustCompile(regexp.QuoteMeta(buildroot()) + `(.*)$`)
         fmt.Println(re.FindStringSubmatch(path))
         if len(re.FindStringSubmatch(path)) > 1 {
           if info.IsDir() {
             list = append(list, "%dir " + re.FindStringSubmatch(path)[1])
           } else {
             list = append(list, re.FindStringSubmatch(path)[1])
           }
         }
         return nil
       })
  check(e)

  f, e := os.Create("file.lst")
  check(e)
  defer f.Close()

  for _, s := range list {
    f.WriteString(s + "\n")
  }
}

func godoc() {
  fmt.Println("We should generate proper godocs!")
}

func main() {
  opts := os.Args
  size := len(opts)

  var action,importpath,modifier,extraflags string

  valid := map[string]func() {"arch":arch,
                              "prep":prep,
                              "build":build,
                              "install":install,
                              "source":source,
                              "test":test,
                              "filelist":filelist,
                              "godoc":godoc}

  if size == 1 {
    // print help
    fmt.Println("Please specify a valid metho: arch, prep, build, install, source, test, filelist, godoc")
  }

  if size == 2 {
    action = opts[1]
  }

  if size > 2 {
    action = opts[1]
    if action == "prep" {
      importpath = opts[2]
    }
    if action == "test" || action == "build" {
      _, modifier, extraflags = parse_args(opts[2:])
    }
  }

  paths := map[string]string {"importpath":importpath,
                              "modifier":modifier,
                              "extraflags":extraflags}

  for k, v := range paths {
    if v != "" {
      store(k, v)
    }
  }

  _, ok := valid[action]
  if !ok { panic("Not a valid action") }

  valid[action]()
}
