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

# ARGV[0], the called method itself
if ARGV[0] == "--arch"

	puts $go_arch

elsif ARGV[0] == "--prep"

	puts "Preparation Stage:\n"

	# ARGV[1] the import path
	if ARGV[1] == nil
		puts "[ERROR]Empty IMPORTPATH! Please specify a valid one.\n"
	else
		gopath = $builddir + "/go"
		puts "GOPATH set to: " + gopath + "\n"

		importpath = ARGV[1]
		puts "IMPORTPATH set to: " + importpath + "\n"

		# export IMPORTPATH to a temp file, as ruby can't export system environment variables
                # like shell scripts
		File.open("/tmp/importpath.txt","w:UTF-8") do |f|
			f.puts(importpath)
		end

		# return current directory name, eg: ruby-2.2.4
		dir = File.basename(Dir.pwd)
		destination = gopath + "/src/" + importpath
		puts "Creating " + destination + "\n"
		FileUtils.mkdir_p(destination)

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
	end

	puts "Preparation Finished!\n"

elsif ARGV[0] == "--build"

	puts "Build stage:\n"

	gopath = $builddir + "/go:" + $libdir + "/go/contrib"
        gobin = $builddir + "/go/bin"
	buildflags = "-s -v -p 4 -x"

	# get importpath from /tmp/importpath.txt saved by prep()
	importpath = open("/tmp/importpath.txt","r:UTF-8").gets.strip!

	opts = Opts.get_opts
	mods = Opts.get_mods
	extraflags = ""

	unless opts.empty?
		if opts.include?("--with-buildid")
			buildid = "0x" + SecureRandom.hex(20)
			# whitespace is important!
			extraflags = extraflags + ' -ldflags "-B ' + buildid + '"'
			opts.delete("--with-buildid")
		end

		opts.each do |o|
			extraflags = extraflags + " #{o}"
		end
	end

	# MODs: nil, "...", "/...", "foo...", "foo/...", "foo bar", "foo bar... baz" and etc
	if mods.empty?
		CLI.run({"GOPATH"=>gopath,"GOBIN"=>gobin}, "go install #{extraflags} #{buildflags} #{importpath}")	
	else
		for mod in mods do
			if mod == "..."
				CLI.run({"GOPATH"=>gopath,"GOBIN"=>gobin}, "go install #{extraflags} #{buildflags} #{importpath}...")
				break
			else
				CLI.run({"GOPATH"=>gopath,"GOBIN"=>gobin}, "go install #{extraflags} #{buildflags} #{importpath}/#{mod}")
			end
		end
	end

	puts "Build Finished!\n"

elsif ARGV[0] == "--install"

	puts "Installation stage:\n"

	# check exitstatus
        File.open("/tmp/exitstatus.txt","r:UTF-8") {|f| abort "Previous stage failed! Abort!" if f.read != "0\n" }

	unless Dir["#{$builddir}/go/pkg/*"].empty?
		puts "Copying generated stuff to " + $buildroot + $go_contribdir
		Dir.glob($builddir + "/go/pkg/linux_" + $go_arch + "/*").each do |f|
			puts "Copying " + f
			FileUtils.cp_r(f, $buildroot + $go_contribdir)
		end
		puts "Done!"
	end

	unless Dir["#{$builddir}/go/bin/*"].empty?
		puts "Copyig binaries to " + $buildroot + "/usr/bin"
		Dir.glob($builddir + "/go/bin/*").each do |f|
			puts "Copying " + f
			FileUtils.chmod_R(0755,f)
			FileUtils.cp_r(f,$buildroot + "/usr/bin")
		end
		puts "Done!"
	end

	puts "Install finished!\n"

elsif ARGV[0] == "--source"

	puts "Source package creation:"

	# check exitstatus
        File.open("/tmp/exitstatus.txt","r:UTF-8") {|f| abort "Previous stage failed! Abort!" if f.read != "0\n" }

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

	puts "Source package created!"

elsif ARGV[0] == "--fix"

	puts "Fixing stuff..."

	# check exitstatus
        File.open("/tmp/exitstatus.txt","r:UTF-8") {|f| abort "Previous stage failed! Abort!" if f.read != "0\n" }

        # only "--fix" is given, no other parameters
        if ARGV.length == 1
                puts "[ERROR]gofix: please specify a valid importpath, see: go help fix"
        else
                gopath = $builddir + "/go"
                CLI.run({"GOPATH"=>gopath},"go fix #{ARGV[1]}...")
        end

	puts "Fixed!"

elsif ARGV[0] == "--test"

	puts "Testing codes..."

	# check exitstatus
        File.open("/tmp/exitstatus.txt","r:UTF-8") {|f| abort "Previous stage failed! Abort!" if f.read != "0\n" }

	# only "--test" is given, no other parameters
	if ARGV.length == 1
		puts "[ERROR]gotest: please specify a valid importpath, see: go help test"
	else
		gopath = $builddir + "/go:" + $libdir + "/go/contrib"
		CLI.run({"GOPATH"=>gopath}, "go test -x #{ARGV[1]}...")
	end

	puts "Test passed!"

# generate filelist for go_contribdir go_contribdir_dir
elsif ARGV[0] == "--filelist"

	puts "Processing filelists..."

	# check exitstatus
        File.open("/tmp/exitstatus.txt","r:UTF-8") {|f| abort "Previous stage failed! Abort!" if f.read != "0\n" }

	opts = Opts.get_opts
	excludes = Opts.get_mods[0]
	# two directories, one is /BUILD/go. reject it, returns array with 1 element.
	builddirs = Dir.glob($builddir + "/*").reject! { |f| f == $builddir + "/go" }
	builddir = builddirs[0]

	# find shared build from linux_amd64_dynlink
	if opts.include?("--shared")
		Filelists.new($buildroot + $go_contribdir + "_dynlink", builddir + "/shared.lst")
		Filelists.new($buildroot + "/usr/bin", builddir + "/shared.lst")
		Filelists.new($buildroot + $go_tooldir,builddir + "/shared.lst")
		Filelists.exclude(builddir + "/shared.lst",excludes)
	# process for -source sub-package
	elsif opts.include?("--source")
		Filelists.new($buildroot + $go_contribsrcdir, builddir + "/source.lst")
		Filelists.exclude(builddir + "/source.lst",excludes)
	# default for main package, static build
	else
		Filelists.new($buildroot + $go_contribdir,builddir + "/file.lst")
		Filelists.new($buildroot + "/usr/bin", builddir + "/file.lst")
		Filelists.new($buildroot + $go_tooldir,builddir + "/file.lst")
		Filelists.exclude(builddir + "/file.lst",excludes)
	end

	puts "Filelists created!"

else

	puts "Please specify a valid method: --prep, --build, --install, --fix, --test, --source, --filelist"

end

