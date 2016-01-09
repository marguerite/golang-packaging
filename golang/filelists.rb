module Filelists

	require 'find'
	require 'fileutils'
	require File.join(File.dirname(__FILE__),'rpmsysinfo.rb')
	include RpmSysinfo

	@@buildroot = RpmSysinfo.get_buildroot
	@@contribdir = RpmSysinfo.get_go_contribdir
	@@contribsrcdir = RpmSysinfo.get_go_contribsrcdir
	@@tooldir = RpmSysinfo.get_go_tooldir
	@@bindir = "/usr/bin"
	@@outfile = ""

	def self.new(path,outfile)

		@@outfile = outfile

		File.open(@@outfile,"a:UTF-8") do |f|

			Find.find(path) do |f1|

				# ignore ".gitignore stuff" but let importpath like
				# "code.google.com" go
				unless f1.gsub(/^.*\//,'').index(/^\./)
					# %file section doesn't need buildroot
					f2 = f1.gsub(@@buildroot,"")
					unless [@@contribdir, @@contribsrcdir, @@tooldir, @@bindir].include?(f2)
						if File.directory?(f1)
							f.puts("%dir " + f2)
						else
							f.puts(f2)
						end
					end
				end

			end			

		end

	end

	def exclude(infile=@@outfile,excludes=[])

        	# treat excludes as regex array, eg:
        	# ["/usr/bin/*","/usr/lib/debug/*"]
		if excludes == nil
			a = Array.new
		else
			a = excludes.split("\s")
		end

        	# expand the regex to actual file
        	list = Array.new
        	a.each do |i|

                	Dir.glob(@@buildroot + i) do |f|

                        	list << f.gsub(@@buildroot,"")

                	end

        	end

		# delete the files excluded from buildroot
		list.each { |f| File.delete(@@buildroot + f) }

        	File.open(infile,"r:UTF-8") do |f|

                	File.open(infile + ".new","w:UTF-8") do |f1|

                        	f.each_line do |l|

                                	f1.puts(l) unless list.include?(l.chomp!)

                        	end

                	end

        	end

		FileUtils.mv(infile + ".new",infile)

	end

end

