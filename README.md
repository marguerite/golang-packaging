[![Go Report Card](https://goreportcard.com/badge/github.com/marguerite/golang-packaging)](https://goreportcard.com/report/github.com/marguerite/golang-packaging)

golang-packaging

------

RPM macros and utilities for golang packaging.

A golang packager can just say

>  BuildRequires: golang-packaging 

and let the included scripts handle Provides/Requires for
you automatically. You can also say 

>  BuildRequires: golang(xxx) 

in specifications for packages built with golang-packaging.

Important Changes in v2:

* rewritten in golang itself from scrach.
* Support golang 1.16, that is, now golang-packaging can automatically detect if the build is in GOPATH mode or Module-aware mode. For Module-aware mode, vendor directory is not respected, but it is still useful: we use the BuildRequires and the vendor directory to fill up go/pkg/mod directory.
In Module-aware mod, golang will check the zip hash of the importpath directory with version, eg the sha1 hash of the zip archive of "github.com/marguerite/golang-packaging@v16.0.0", and the content hash of go.mod inside that importpath, eg the sha1 hash of the content of go.mod inside "github.com/marguerite/golang-packaging@v16.0.0", against the hashes in go.sum of the dependebt. so we fake such hashes, even the files themselves too to make use of BuildRequires and vendor directory.
Such implementation needs lot of programming, so bash is no more.
* In prepare step, we no longer hard copy RPM_BUILD_DIR and go contrib src directory, symlink make the build faster
* Use "runtime.GOARCH" instead of "uname -m", since the build happens in virtual environments, "uname -m" may occasionaly use the host's architecture so "-buildmode=pie" may be occasionally appended on x86_64, in that case a full rebuild of golang itself will happen.
