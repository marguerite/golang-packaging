# Macros for Go module building.
#
# Copyright: (c) 2011 Sascha Peilicke <saschpe@gmx.de>
# Copyright: (c) 2012 Graham Anderson <graham@andtech.eu>
# Copyright: (c) 2015 SUSE Linux GmbH
#


%go_ver           %(LC_ALL=C rpm -q --qf '%%{epoch}:%%{version}\\n' go | sed -e 's/(none)://' -e 's/ 0:/ /' | grep -v "is not")
%go_arch	        %(%{_prefix}/lib/rpm/golang.sh arch)
%go_build_ver     %(go version | sed 's/^go version //' | sed 's:\/::g' | tr -d ' ' | cut -c 1-7 )
%go_api_ver       %(echo %{go_ver} | sed 's/\.[0-9]$//')

%go_dir           %{_libdir}/go
%go_bindir        %{_libdir}/go/bin
%go_srcdir        %{_libdir}/go/src
%go_sitedir       %{_libdir}/go/pkg
%go_sitearch      %{_libdir}/go/pkg/linux_%{go_arch}
%go_contribdir    %{_libdir}/go/contrib/pkg/linux_%{go_arch}
%go_contribsrcdir %{_datadir}/go/contrib/src/
%go_tooldir       %{_datadir}/go/pkg/tool/linux_%{go_arch}

%go_nostrip \
%undefine _build_create_debug \
%define __arch_install_post export NO_BRP_STRIP_DEBUG=true

%go_exclusivearch \
ExclusiveArch: aarch64 %ix86 x86_64 %arm ppc64 ppc64le s390x

%go_provides \
%if 0%{?suse_version} <= 1110 \
%global _use_internal_dependency_generator 0 \
%global __find_provides %{_prefix}/lib/rpm/golang.prov \
%global __find_requires %{_prefix}/lib/rpm/golang.req \
%endif

# goprep prepares the expected Go package build environement. We need a $GOPATH
# (for reference look at go help gopath) and we need a valid importpath (for
# reference look at go help packages)
%goprep \
%{_prefix}/lib/rpm/golang.sh prep

# gobuild macro actually performs the command "go install", but the go toolchain
# will install to the $GOPATH which allows us then customise the final install
# for the distro default locations.
%gobuild \
%{_prefix}/lib/rpm/golang.sh build

# goinstall moves the binary files into the bin folder, don't mix it with the go
# install command since this really just copies files and doesn't execute
# anything else.
%goinstall \
%{_prefix}/lib/rpm/golang.sh install

# gosrc copies over all source files into the contrib source directory to be on
# a properly packaged location.
%gosrc \
%{_prefix}/lib/rpm/golang.sh source

# gotest can execute the integrated test suite to make sure the software really
# works like expected in our environment.
%gotest \
%{_prefix}/lib/rpm/golang.sh test

# godoc should generate useable documentations based on the inline godoc
# comments of the source files.
%godoc \
%{_prefix}/lib/rpm/golang.sh godoc

# go_filelist generates different lists of files to be consumed by the file
# section of an rpm.
%gofilelist \
%{_prefix}/lib/rpm/golang.sh filelist
