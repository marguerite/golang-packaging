module Filelists

	require 'find'
	require File.join(File.dirname(__FILE__),'rpmsysinfo.rb')
	include RpmSysinfo

	@@buildroot = RpmSysinfo.get_buildroot
	@@contribdir = RpmSysinfo.get_go_contribdir
	@@contribsrcdir = RpmSysinfo.get_go_contribsrcdir
	@@tooldir = RpmSysinfo.get_go_tooldir
	@@bindir = "/usr/bin"

	def self.new(path,outfile)

		File.open(outfile,"a:UTF-8") do |f|

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

end

