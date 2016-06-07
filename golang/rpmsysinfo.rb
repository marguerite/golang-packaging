module RpmSysinfo

	require 'rbconfig'

	# x86_64-(gnu|linux|blabla...)
	@@rbarch = RUBY_PLATFORM.gsub(/-.*$/,"")
	# architectures are defined in /usr/lib/rpm/macros
	@@ix86 = ["i386","i486","i586","i686","pentium3","pentium4","athlon","geode"]
	@@arm = ["armv3l","armv4b","armv4l","armv4tl","armv5b","armv5l","armv5teb","armv5tel","armv5tejl","armv6l","armv6hl","armv7l","armv7hl","armv7hnl"]
	if @@ix86.include?(@@rbarch)
		@@go_arch = "386"
	end
	if @@rbarch == "x86_64"
        	@@go_arch = "amd64"
	end
	if @@arm.include?(@@rbarch)
        	@@go_arch = "arm"
	end
	if @@rbarch == "aarch64"
		@@go_arch = "arm64"
	end
	if @@rbarch == "powerpc64"
		@@go_arch = "ppc64"
	end
	if @@rbarch == "powerpc64le"
		@@go_arch = "ppc64le"
	end

	def self.get_builddir

		return ENV["RPM_BUILD_DIR"]

	end

	def self.get_buildroot

		return ENV["RPM_BUILD_ROOT"]

	end

	def self.get_libdir

		return RbConfig::CONFIG['libdir']

	end

	def self.get_go_arch

		return @@go_arch

	end

	def self.get_go_contribdir

		return self.get_libdir + "/go/contrib/pkg/linux_" + @@go_arch

	end

	def self.get_go_tooldir

		return "/usr/share/go/pkg/tool/linux_" + @@go_arch

	end

	def self.get_go_contribsrcdir

		return "/usr/share/go/contrib/src"
	
	end

	def self.get_go_importpath

		# this funtion is used in golang.prov/req only
		# called after %go_prep
		# so the simplest method is to read /tmp/importpath.txt
		return open("/tmp/importpath.txt","r:UTF-8").gets.strip!        

	end

end

