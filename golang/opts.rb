module Opts

	@@inputs = ARGV

	@@opts, @@mods = [],[]

	# ARGV: -buildmode=shared -linkshared -cflags="a b c" --with-buildid -tag tag d e...
	@@inputs.each do |i|

		# options begin with "-"
		if i.index(/^-/)

			if i.index("-tags")

				# the following ARG is actually for -tag, not the main script

				@@opts << i + "\s"+ @@inputs[@@inputs.index(i) + 1]

				@@inputs.delete_at(@@inputs.index(i) + 1)

			else

				@@opts << i

			end

		else

			@@mods << i

		end

	end

	def self.get_opts		

		# remove the first opt which is "--build, --prep" and etc
		if @@opts.length > 1
			return @@opts.reject! {|f| f == ARGV[0] }
		else
			return []
		end

	end

	def self.get_mods

		return @@mods

	end

end

