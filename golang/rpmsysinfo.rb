module RpmSysinfo

	if File.directory?("/usr/src/packages") & File.writable?("/usr/src/packages")
        	@@topdir = "/usr/src/packages"
	else
        	@@topdir = ENV["HOME"] + "/rpmbuild"
	end

        @@buildroot = Dir.glob(@@topdir + "/BUILDROOT/*")[0]

        # sometimes buildroot locates in tmppath/name-version-build

        @@buildroot = Dir.glob("/var/tmp/*-build")[0] if @@buildroot == nil

	@@archdir = Dir.glob(@@buildroot + "/usr/lib*/go/contrib/pkg/*")[0] + "/"

	@@specfile = Dir.glob(@@topdir+ "/SOURCES/*.spec")[0]


	def self.set_topdir(top)

		@@topdir = top

	end

	def self.get_topdir

		return @@topdir

	end

	def self.get_buildroot

		return @@buildroot

	end

	def self.get_archdir

		return @@archdir

	end

	def self.get_importpath

		importpath = ""

                File.open(@@specfile) do |f|

                        f.each_line do |l|

                                found = 0

                                # see if there's any packager definition for importpath
                                if l.index(/%(define|global)[\s]+(import_path|importpath)/i) then

                                        importpath = l.gsub(/%(define|global)[\s]+(import_path|importpath)/i,'').lstrip!.chomp!.gsub(/"/,'').gsub(/"/,'')

                                        found = 1

                                end

                                # use the one in "%goprep code.google.com/p/log4go"
                                if (found == 0 && l.index("%goprep")) then 

                                        importpath = l.gsub(/%goprep/,'').lstrip!.chomp!

                                        found = 1

                                end

                                # sometimes packager didn't package using the macros we give, extract from URL tag
                                if (found == 0 && l.index("Url:")) then

                                        # eg: "URL: https://code.google.com/p/log4go/"
                                        # "URL: http://download.fcitx-im.org/fcitx/fcitx"

                                        # gsub
                                        # 1. remove "Url:" then leading whitespace and ending "\n"
                                        # 2. remove "http://" or "https://"
                                        importpath = l.gsub(/Url:/,'').lstrip!.chomp!.gsub(/^(http|https)\:\/\//,'')

                                end

                        end

                end

                # code.google.com/p/log4go/, remove the ending "/"
                importpath = importpath.gsub(/\/$/,'') if importpath.index(/\/$/)
            
                return importpath
        
        end

end

