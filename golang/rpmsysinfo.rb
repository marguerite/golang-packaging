module RpmSysinfo

	if File.directory?("/usr/src/packages") & File.writable?("/usr/src/packages")
        	@@topdir = "/usr/src/packages"
	else
        	@@topdir = ENV["HOME"] + "/rpmbuild"
	end

        @@buildroot = Dir.glob(@@topdir + "/BUILDROOT/*")[0]

        # sometimes buildroot locates in tmppath/name-version-build

        @@buildroot = Dir.glob("/var/tmp/*-build")[0] if @@buildroot == nil

	@@arch = ""
	# x86_64-(gnu|linux|blabla...)
	@@rbarch = RUBY_PLATFORM.gsub(/-.*$/,"")
	# architectures are defined in /usr/lib/rpm/macros
	@@ix86 = ["i386","i486","i586","i686","pentium3","pentium4","athlon","geode"]
	@@arm = ["armv3l","armv4b","armv4l","armv4tl","armv5b","armv5l","armv5teb","armv5tel","armv5tejl","armv6l","armv6hl","armv7l","armv7hl","armv7hnl"]
	if @@ix86.include?(@@rbarch)
		@@libdir = "/usr/lib"
		@@go_arch = "386"
		@@arch = "i386"
	end
	if @@rbarch == "x86_64"
        	@@libdir = "/usr/lib64"
        	@@go_arch = "amd64"
		@@arch = @@rbarch
	end
	if @@arm.include?(@@rbarch)
        	@@libdir = "/usr/lib"
        	@@go_arch = "arm"
		@@arch = @@rbarch
	end
	if @@rbarch == "aarch64"
		@@libdir = "/usr/lib64"
		@@go_arch = "arm64"
		@@arch = @@rbarch
	end
	if @@rbarch == "ppc64"
		@@libdir = "/usr/lib64"
		@@go_arch = "ppc64"
		@@arch = @@rbarch
	end
	if @@rbarch == "ppc64le"
		@@libdir = "/usr/lib64"
		@@go_arch = "ppc64le"
		@@arch = @@rbarch
	end

	def self.set_topdir(top)

		@@topdir = top

	end

	def self.get_topdir

		return @@topdir

	end

	def self.get_builddir

		return @@topdir + "/BUILD"

	end

	def self.get_buildroot

		return @@buildroot

	end

	def self.get_libdir

		return @@libdir

	end

	def self.get_arch

		return @@go_arch

	end

	def self.get_contribdir

		go_contribdir = @@libdir + "/go/contrib/pkg/linux_" + @@go_arch
		return go_contribdir

	end

	def self.get_tooldir

		go_tooldir = "/usr/share/go/pkg/tool/linux_" + @@go_arch
		return go_tooldir

	end

	def self.get_contribsrcdir

		return "/usr/share/go/contrib/src"
	
	end

	def self.get_importpath

		importpath = ""

		specfile = Dir.glob(@@topdir + "/SOURCES/*.spec")[0]

                File.open(specfile) do |f|

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

