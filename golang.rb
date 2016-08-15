#!/usr/bin/env ruby

require 'fileutils'
require 'securerandom'
require 'find'
require File.join(File.dirname(__FILE__),'golang/rpmsysinfo.rb')
require File.join(File.dirname(__FILE__),'golang/opts.rb')
require File.join(File.dirname(__FILE__),'golang/filelists.rb')
require File.join(File.dirname(__FILE__),'golang/cli.rb')
include RpmSysinfo
include Opts
include Filelists
include CLI

# GLOBAL RPM MACROS

$builddir = RpmSysinfo.get_builddir
$buildroot = RpmSysinfo.get_buildroot
$libdir = RpmSysinfo.get_libdir
$go_arch = RpmSysinfo.get_go_arch
$go_contribdir = RpmSysinfo.get_go_contribdir
$go_contribsrcdir = RpmSysinfo.get_go_contribsrcdir
$go_tooldir = RpmSysinfo.get_go_tooldir

def goprep(importpath=nil)
	puts "Preparation Stage:\n"
	if importpath.nil?
		puts "[ERROR]Empty IMPORTPATH! Please specify a valid one.\n"
	else
		gopath = $builddir + "/go"
		# return current directory name, eg: ruby-2.2.4
		dir = File.basename(Dir.pwd)
		destination = gopath + "/src/" + importpath

		puts "GOPATH set to: " + gopath + "\n"

		puts "IMPORTPATH set to: " + importpath + "\n"
		# export IMPORTPATH to a temp file, as ruby can't export system environment variables
	  	# like shell scripts
		open("/tmp/importpath","w:UTF-8") { |f| f.puts(importpath) }

		puts "Creating " + destination + "\n"
		FileUtils.mkdir_p(destination)
		puts "Creating GOBIN directory\n"
		FileUtils.mkdir_p(gopath + "/bin")

		# copy everything to destination
		puts "Copying everything under " + $builddir + "/" + dir + " to " + destination + " :\n"
		Dir.glob($builddir + "/" + dir + "/*").each do |f|
			puts "Copying " + f + "\n"
			FileUtils.cp_r(f, File.join(destination, File.basename(f)))
		end
		puts "Files are moved!\n"

		# create target directories
		puts "Creating directory for binaries " + $buildroot + "/usr/bin" + "\n"
		FileUtils.mkdir_p($buildroot + "/usr/bin")
		puts "Creating directory for contrib " + $buildroot + $go_contribdir + "\n"
		FileUtils.mkdir_p($buildroot + $go_contribdir)
		puts "Creating directory for source " + $buildroot + $go_contribsrcdir + "\n"
		FileUtils.mkdir_p($buildroot + $go_contribsrcdir)
		puts "Creating directory for tool " + $buildroot + $go_tooldir + "\n"
		FileUtils.mkdir_p($buildroot + $go_tooldir)
		puts "Preparation Finished!\n"
	end
end

def gobuild()
	puts "Build stage:\n"
	gopath = $builddir + "/go:" + $libdir + "/go/contrib"
  	gobin = $builddir + "/go/bin"
	buildflags = "-s -v -p 4 -x"
	# get importpath from /tmp/importpath saved by prep()
	importpath = open("/tmp/importpath","r:UTF-8").gets.strip!
	opts = Opts.get_opts
	mods = Opts.get_mods
	extraflags = String.new
	sharedflags = " -buildmode=shared -linkshared "
	cmd = "build"

	unless opts.empty?
		if opts.include?("--with-buildid")
			buildid = "0x" + SecureRandom.hex(20)
			# whitespace is important!
			extraflags = extraflags + ' -ldflags "-B ' + buildid + '"'
			opts.delete("--with-buildid")
		end
		if opts.include?("--enable-shared")
			extraflags = extraflags + sharedflags
			opts.delete("--enable-shared")
			# touch /tmp/shared for indentification of shared build later
			FileUtils.touch "/tmp/shared"
			# change build command to 'go install'
			cmd = "install"
		end
		opts.each do |o|
			extraflags = extraflags + " #{o}"
		end
	end

	# go build will leave binaries in currently directory, save the contents of current directory.
	current = Array.new
	Dir.glob(Dir.pwd + "/*") {|f| current << f}

	# MODs: nil, "...", "/...", "foo...", "foo/...", "foo bar", "foo bar... baz" and etc
	if mods.empty?
		CLI.run({"GOPATH"=>gopath,"GOBIN"=>gobin}, "go #{cmd} #{extraflags} #{buildflags} #{importpath}")
	else
		for mod in mods do
			if mod == "..."
				CLI.run({"GOPATH"=>gopath,"GOBIN"=>gobin}, "go #{cmd} #{extraflags} #{buildflags} #{importpath}")
				Dir.glob($builddir + "/go/src/" + importpath + "/**/*") do |d|
					if File.directory?(d)
						buildable = false
						Dir.glob(d + "/*") do |f|
							if f.index(/\.go$/)
								buildable = true
								break
							end
						end
						CLI.run({"GOPATH"=>gopath,"GOBIN"=>gobin}, "go #{cmd} #{extraflags} #{buildflags} " + d.gsub($builddir + "/go/src/",'')) if buildable
					end
				end
				break
			else
				CLI.run({"GOPATH"=>gopath,"GOBIN"=>gobin}, "go #{cmd} #{extraflags} #{buildflags} #{importpath}/#{mod}")
			end
		end
	end

	# go build will leave binaries in current directory, copy them to $GOBIN
	currentnew = Array.new
	Dir.glob(Dir.pwd + "/*") {|f| currentnew << f}
	diff = currentnew - current
	diff.each do |f|
		p f
		FileUtils.install(f,gobin + "/")
	end

	puts "Build Finished!\n"
end

def goinstall()
	puts "Installation stage:\n"
	# check previous exitstatus
	abort "Build stage failed! Abort!" if File.exist?("/tmp/failed")

	importpath = open("/tmp/importpath","r:UTF-8").gets.strip!

	# static: if there's any binary inside $GOBIN, which indicates there's a main function somewhere inside the source code,
	# 	  then it is a Go program that we don't need to distribute its source codes to work as a build dependency.
	#	  then we make it optional through gosource()
	# shared: we'll distribute its built codes (libraries)
	if File.exist? "/tmp/shared"
		unless Dir[$builddir + "/go/pkg/*"].empty?
		# TODO: shared stuff
		end
	elsif Dir[$builddir + "/go/bin/*"].empty?
		puts "This will copy all *.go files in #{$builddir}/go/src, but resource files needed are still not copyed"
		Find.find($builddir + "/go/src") do |f|
                	unless FileTest.directory?(f)
				# don't install test files
                        	if f.index(/\.go$/) && ! f.gsub(importpath,'').index(/(example|test)/)
                                	puts "Copying " + f
                                	FileUtils.chmod_R(0644,f)
                                	# create the same hierarchy
                                	dir = $buildroot + $go_contribsrcdir + f.gsub($builddir + "/go/src",'')
                                	dir1 = dir.gsub(File.basename(dir),'')
                                	FileUtils.mkdir_p(dir1)
                                	FileUtils.install(f,dir1)
				end
			end
		end
		puts "Done!"
	end

	unless Dir[$builddir + "/go/bin/*"].empty?
		puts "Copyig golang binaries to " + $buildroot + "/usr/bin"
		Dir.glob($builddir + "/go/bin/*").each do |f|
			puts "Copying binary" + f
			FileUtils.chmod_R(0755,f)
			FileUtils.cp_r(f,$buildroot + "/usr/bin")
		end
		puts "Done!"
	end

	puts "Install finished!\n"
end

def gosource()
	puts "Source package creation:"
	# check previous exitstatus
	abort "Install stage failed! Abort!" if File.exist?("/tmp/failed")

	if ! Dir[$builddir + "/go/bin/*"].empty? || File.exist?("/tmp/shared")
		puts "This will copy all *.go files in #{$builddir}/go/src, but resource files needed are still not copyed"
		Find.find($builddir + "/go/src") do |f|
			unless FileTest.directory?(f)
				if f.index(/\.go$/)
					puts "Copying " + f
					FileUtils.chmod_R(0644,f)
					# create the same hierarchy
					dir = $buildroot + $go_contribsrcdir + f.gsub($builddir + "/go/src",'')
					dir1 = dir.gsub(File.basename(dir),'')
					FileUtils.mkdir_p(dir1)
					FileUtils.install(f,dir1)
				end
			end
		end
	else
		puts "gosource is only meaningful for Go programs that have binaries inside /usr/bin,"
		puts "or shared builds that have .shlibname suffix. because currently go source codes"
		puts "are distributed by default for static builds. consider remove it from specfile."
	end

	puts "Source package created!"
end

def gofix()
	puts "Fixing stuff..."
	# check previous exitstatus
	abort "Previous stage failed! Abort!" if File.exist?("/tmp/failed")
	# "--fix" should be given without other parameters
	if ARGV.length <= 1
		puts "[ERROR]gofix: please specify a valid importpath, see: go help fix"
	else
		gopath = $builddir + "/go"
		CLI.run({"GOPATH"=>gopath},"go fix #{ARGV[1]}...")
	end
	puts "Fixed!"
end

def gotest()
	puts "Testing codes..."
	# check previous exitstatus
	abort "Previous stage failed! Abort!" if File.exist?("/tmp/failed")
	# "--test" should be given without other parameters
	if ARGV.length <= 1
		puts "[ERROR]gotest: please specify a valid importpath, see: go help test"
	else
		gopath = $builddir + "/go:" + $libdir + "/go/contrib"
		CLI.run({"GOPATH"=>gopath}, "go test -x #{ARGV[1]}...")
	end
	puts "Test passed!"
end

def gofiles()
	puts "Processing filelists..."
	# check previous exitstatus
	abort "Previous stage failed! Abort!" if File.exist?("/tmp/failed")

	opts = Opts.get_opts
	excludes = Opts.get_mods[0]
	# two directories, one is /BUILD/go. reject it, returns array with 1 element.
	builddirs = Dir.glob($builddir + "/*").reject! { |f| f == $builddir + "/go" }
	builddir = builddirs[0]

	# find shared build from linux_amd64_dynlink
	if opts.include?("--enable-shared")
		Filelists.new($buildroot + $go_contribdir + "_dynlink", builddir + "/shared.lst")
		Filelists.new($buildroot + "/usr/bin", builddir + "/shared.lst")
		Filelists.new($buildroot + $go_tooldir,builddir + "/shared.lst")
		Filelists.exclude(builddir + "/shared.lst",excludes)
	# process for -source sub-package
	elsif opts.include?("--source")
		if ! Dir[$builddir + "/go/bin/*"].empty? || File.exist?("/tmp/shared")
			Filelists.new($buildroot + $go_contribsrcdir, builddir + "/source.lst")
			Filelists.exclude(builddir + "/source.lst",excludes)
		else
                	puts "'go_filelist --source' is only meaningful for Go programs that have binaries inside /usr/bin,"
               		puts "or shared builds that have .shlibname suffix. because currently go source codes"
                	puts "are distributed by default for static builds. consider remove it from specfile."
		end
	# default for main package, static build
	else
		if Dir[$builddir + "/go/bin/*"].empty? 
			Filelists.new($buildroot + $go_contribsrcdir,builddir + "/file.lst")
		end
		Filelists.new($buildroot + "/usr/bin", builddir + "/file.lst")
		Filelists.new($buildroot + $go_tooldir,builddir + "/file.lst")
		Filelists.exclude(builddir + "/file.lst",excludes)
	end
	puts "Filelists created!"
end

# ARGV[0], the called method itself
if ARGV[0] == "--arch"
	puts $go_arch
elsif ARGV[0] == "--prep"
	# ARGV[1] the import path or nil
	goprep(ARGV[1])
elsif ARGV[0] == "--build"
	gobuild()
elsif ARGV[0] == "--install"
	goinstall()
elsif ARGV[0] == "--source"
	gosource()
elsif ARGV[0] == "--fix"
	gofix()
elsif ARGV[0] == "--test"
	gotest()
# generate filelist for go_contribdir go_contribdir_dir
elsif ARGV[0] == "--filelist"
	gofiles()
else
	puts "Please specify a valid method: --prep, --build, --install, --fix, --test, --source, --filelist"
end
