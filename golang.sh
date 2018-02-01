#!/usr/bin/env bash
set -e
set -x

store_buildroot_path() {
  echo ${RPM_BUILD_ROOT} >| /tmp/buildrootpath.txt
}

store_import_path() {
  echo ${1} >| /tmp/importpath.txt
}

check_import_path() {
  if [[ "$(get_import_path)" == "" ]]; then
    echo "Empty import path, please specify a valid one!" >&2
    exit 1
  fi
}

get_import_path() {
  head /tmp/importpath.txt
}

get_build_path() {
  echo "${RPM_BUILD_DIR}/go"
}

get_buildcontrib_path() {
  echo "${RPM_BUILD_DIR}/contrib"
}

get_gobin_path() {
  echo "${RPM_BUILD_DIR}/go/bin"
}

get_gocontrib_path() {
  echo "$(rpm --eval %{_datadir})/go/$(rpm --eval %go_api_ver)/contrib"
}

get_contrib_path() {
  echo $(rpm --eval %go_contribdir)
}

get_source_path() {
  echo $(rpm --eval %go_contribsrcdir)
}

get_tool_path() {
  echo $(rpm --eval %go_tooldir)
}

get_binary_path() {
  echo $(rpm --eval %_bindir)
}

get_destination_path() {
  echo "$(get_build_path)/src/$(get_import_path)"
}

process_arch() {
  local arch=$(uname -m)

  case "${arch}" in
  "x86_64")
    echo "amd64"
    ;;
  "i386"|"i486"|"i586"|"i686"|"pentium3"|"pentium4"|"athlon"|"geode")
    echo "386"
    ;;
  aarch64*)
    echo "arm64"
    ;;
  armv*)
    echo "arm"
    ;;
  *)
    echo $arch
    ;;
  esac
}

process_prepare() {
  store_import_path "${1}"
  check_import_path
  store_buildroot_path

  echo "Creating build path $(get_destination_path)"
  rm -rf $(get_destination_path)
  mkdir -p $(get_destination_path)

  echo "Creating deps path $(get_buildcontrib_path)/src"
  rm -rf $(get_buildcontrib_path)/src
  mkdir -p $(get_buildcontrib_path)/src

  echo "Copying files to $(get_destination_path)"
  cp -rT $(pwd) $(get_destination_path)/

  echo "Copying deps to $(get_buildcontrib_path)"
  cp -rT $(get_gocontrib_path)/src $(get_buildcontrib_path)/src

  local contrib_dir=${RPM_BUILD_ROOT}$(get_contrib_path)
  echo "Creating contrib path ${contrib_dir}"
  mkdir -p ${contrib_dir}

  local source_dir=${RPM_BUILD_ROOT}$(get_source_path)
  echo "Creating source path ${source_dir}"
  mkdir -p ${source_dir}

  local tool_dir=${RPM_BUILD_ROOT}$(get_tool_path)
  echo "Creating tool path ${tool_dir}"
  mkdir -p ${tool_dir}

  local binary_dir=${RPM_BUILD_ROOT}$(get_binary_path)
  echo "Creating binary path ${binary_dir}"
  mkdir -p ${binary_dir}
}

process_build() {
  check_import_path


  if [[ "$#" -eq 0 ]]; then
    local modifier=
    local last=0
  else
    local modifier="${@: -1}"
    local last=$(($#-1))
  fi

  local build_flags="-s -v -p 4 -x -buildmode=pie"
  local extra_flags=(
    "${@:1:$last}"
  )

  case "${modifier}" in
  "...")
    GOPATH=$(get_build_path):$(get_buildcontrib_path) GOBIN=$(get_gobin_path) go \
      install ${build_flags} "${extra_flags[@]}" $(get_import_path)...
    ;;
  "")
    GOPATH=$(get_build_path):$(get_buildcontrib_path) GOBIN=$(get_gobin_path) go \
      install ${build_flags} "${extra_flags[@]}" $(get_import_path)
    ;;
  *)
    GOPATH=$(get_build_path):$(get_buildcontrib_path) GOBIN=$(get_gobin_path) go \
      install ${build_flags} "${extra_flags[@]}" $(get_import_path)/${modifier}
    ;;
  esac
}

process_install() {
  check_import_path

  for file in $(find $(get_gobin_path) -type f); do
    echo "Copying $(basename ${file}) to ${RPM_BUILD_ROOT}$(get_binary_path)"
    install -D -m0755 ${file} ${RPM_BUILD_ROOT}$(get_binary_path)
  done
}

process_source() {
  echo "This will copy all *.go, *.s and *.h files in $(get_build_path)/src without resources"

  for file in $(find $(get_build_path)/src -type f -iname \*.go); do
    local destination=${RPM_BUILD_ROOT}$(get_source_path)${file#$(get_build_path)/src/}

    echo "Copying ${file} to ${destination}"
    install -D -m0644 ${file} ${destination}
  done

  for file in $(find $(get_build_path)/src -type f -iname \*.s); do
    local destination=${RPM_BUILD_ROOT}$(get_source_path)${file#$(get_build_path)/src/}

    echo "Copying ${file} to ${destination}"
    install -D -m0644 ${file} ${destination}
  done

  for file in $(find $(get_build_path)/src -type f -iname \*.h); do
    local destination=${RPM_BUILD_ROOT}$(get_source_path)${file#$(get_build_path)/src/}

    echo "Copying ${file} to ${destination}"
    install -D -m0644 ${file} ${destination}
  done
}

process_test() {
  if [[ "${1}" == "" ]]; then
    echo "Please specify a valid importpath, reference: go help test" >&2
    exit 1
  fi

  local modifier="${@: -1}"

  if [[ "$#" -eq 0 ]]; then
    local last=0
  else
    local last=$(($#-1))
  fi

  local extra_flags=(
    "${@:1:$last}"
  )

  GOPATH=$(get_build_path):$(get_buildcontrib_path) GOBIN=$(get_gobin_path) go \
    test "${extra_flags[@]}" -x ${modifier}
}

process_filelist() {
  local file_list="file.lst"

  rm -f ${file_list}

  for path in $(find ${RPM_BUILD_ROOT}$(get_source_path)); do
    local destination=${path#${RPM_BUILD_ROOT}}

    if [[ -d ${path} ]]; then
      echo "%dir ${destination}" >> ${file_list}
    else
      echo "${destination}" >> ${file_list}
    fi
  done
}

process_godoc() {
  echo "We should generate proper godocs!"
}

main() {
  local action="${1}"

  case "${action}" in
  "--arch"|"arch")
    process_arch "${@:2}"
    ;;
  "--prep"|"prep")
    process_prepare "${@:2}"
    ;;
  "--build"|"build")
    process_build "${@:2}"
    ;;
  "--install"|"install")
    process_install "${@:2}"
    ;;
  "--source"|"source")
    process_source "${@:2}"
    ;;
  "--test"|"test")
    process_test "${@:2}"
    ;;
  "--filelist"|"filelist")
    process_filelist "${@:2}"
    ;;
  "--godoc"|"godoc")
    process_godoc "${@:2}"
    ;;
  *)
    echo "Please specify a valid method: arch, prep, build, install, source, test, filelist, godoc" >&2
    ;;
  esac
}

main "$@"
