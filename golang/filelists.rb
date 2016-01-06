module Filelists

	require 'find'
	require File.join(File.dirname(__FILE__),'rpmsysinfo.rb')
	include RpmSysinfo

	def self.new(path,outfile)

		File.open(outfile,"a:UTF-8") do |f|

			Find.find(path) do |f1|

				# ignore ".gitignore stuff" but let importpath like
				# "code.google.com" go
				unless f1.gsub(/^.*\//,'').index(/^\./)
					# %file section doesn't need buildroot
					buildroot = RpmSysinfo.get_buildroot
					f.puts(f1.gsub(buildroot,""))
				end

			end			

		end

	end

end

