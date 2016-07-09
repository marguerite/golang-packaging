# Macros for Go module building.
#
# Copyright: (c) 2011 Sascha Peilicke <saschpe@gmx.de>
# Copyright: (c) 2012 Graham Anderson <graham@andtech.eu>
# Copyright: (c) 2015 SUSE Linux GmbH
#


%go_ver         %(LC_ALL=C rpm -q --qf '%%{epoch}:%%{version}\\n' go | sed -e 's/(none)://' -e 's/ 0:/ /' | grep -v "is not")
%go_arch	GOARCH
%go_build_ver   %(go version | sed 's/^go version //' | sed 's:\/::g' | tr -d ' ' | cut -c 1-7 )
%go_api_ver %(echo %{go_ver} | sed 's/\.[0-9]$//')

%go_dir          %{_libdir}/go
%go_sitedir      %{_libdir}/go/pkg
%go_sitearch     %{_libdir}/go/pkg/linux_%{go_arch}
%go_contribdir     %{_libdir}/go/contrib/pkg/linux_%{go_arch}
%go_contribsrcdir  %{_datadir}/go/contrib/src/
%go_tooldir        %{_datadir}/go/pkg/tool/linux_%{go_arch}

%go_nostrip \
%undefine _build_create_debug \
%define __arch_install_post export NO_BRP_STRIP_DEBUG=true

%go_exclusivearch \
ExclusiveArch:  aarch64 %ix86 x86_64 %arm ppc64 ppc64le s390x

%go_provides \
%if 0%{?suse_version} <= 1110 \
%global _use_internal_dependency_generator 0 \
%global __find_provides %{_prefix}/lib/rpm/golang.prov \
%global __find_requires %{_prefix}/lib/rpm/golang.req \
%endif \
%go_exclusivearch \
Provides:       %{name}-devel = %{version} \
Provides:       %{name}-devel-static = %{version}

# Prepare the expected Go package build environement.
# We need a $GOPATH: go help gopath
# We need a valid importpath: go help packages
%goprep %{_prefix}/lib/rpm/golang.rb --prep

# %%gobuild macro actually performs the command "go install", but the go
# toolchain will install to the $GOPATH which allows us then customise the final
# install for the distro default locations.
#
# gobuild accepts zero or more arguments. Each argument corresponds to
# a modifier of the importpath. If no arguments are passed, this is equivalent
# to the following go install statement:
#
#     go install [importpath]
#
# Only the first or last arguement may be ONLY the wildcard argument "..."
# if the wildcard argument is passed then the importpath expands to all packages
# and binaries underneath it. If the argument contains only the wildcard no further
# arguments are considered.
#
# If no wildcard argument is passed, go install will be invoked on each $arg
# subdirectory under the importpath.
#
# Valid importpath modifier examples:
#
#    example:  %gobuild ...
#    command:  go install importpath...
#
#    example:  %gobuild /...
#    command:  go install importpath/...      (All subdirs NOT including importpath)
#
#    example:  %gobuild foo...
#    command:  go install importpath/foo...   (All subdirs INCLUDING foo)
#
#    example:  %gobuild foo ...               (same as foo...)
#    command:  go install importpath/foo...   (All subdirs INCLUDING foo)
#
#    example:  %gobuild foo/...
#    commands: go install importpath/foo/...  (All subdirs NOT including foo)
#
#    example:  %gobuild foo bar
#    commands: go install importpath/foo
#              go install importpath/bar
#
#    example:  %gobuild foo ... bar
#    commands: go install importpath/foo...   (bar is ignored)
#
#    example:  %gobuild foo bar... baz
#    commands: go install importpath/foo
#              go install importpath/bar...
#              go install importpath/baz
#
# See: go help install, go help packages
%gobuild %{_prefix}/lib/rpm/golang.rb --build

# Install all compiled packages and binaries to the buildroot
%goinstall %{_prefix}/lib/rpm/golang.rb --install

%gofix %{_prefix}/lib/rpm/golang.rb --fix

%gotest %{_prefix}/lib/rpm/golang.rb --test

%gosrc %{_prefix}/lib/rpm/golang.rb --source

%go_filelist %{_prefix}/lib/rpm/golang.rb --filelist

# Template for source sub-package
%gosrc_package(n:r:) \
%package %{-n:-n %{-n*}-}source \
Summary: Source codes for package %{name} \
Group: Development/Sources \
Requires: %{-n:%{-n*}}%{!-n:%{name}} = %{version} \
%{-r:Requires: %{-r*}} \
Provides: %{-n:%{-n*}}%{!-n:%{name}}-doc = %{version}-%{release} \
Obsoletes: %{-n:%{-n*}}%{!-n:%{name}}-doc < %{version}-%{release} \
%description %{-n:-n %{-n*}-}source \
This package provides source codes for package %{name}.\
%{nil}

# backward compatibility
%go_requires    \
%(if [ ! -f /usr/lib/rpm/golang.attr ] ; then \
echo "Requires: go >= %go_build_ver" \
fi) \
%{nil}

%go_recommends %{nil}

%godoc \
%gosrc \
%{nil}

# Template for doc sub-package
%godoc_package(n:r:) \
%package %{-n:-n %{-n*}-}doc \
Summary: API documention for package %{name} \
Group: Documentation/Other \
Requires: %{-n:%{-n*}}%{!-n:%{name}} = %{version} \
%{-r:Requires: %{-r*}} \
%description %{-n:-n %{-n*}-}doc \
This package provides API, examples and documentation \
for package %{name}.\
%{nil}
