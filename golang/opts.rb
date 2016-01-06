module Opts

	@@inputs = ARGV

	@@opts, @@mods = [],[]

	# ARGV: -buildmode=shared -linkshared -cflags="a b c" --with-buildid -tag tag d e...
	@@inputs.each do |i|

		# options begin with "-"
		if i.index(/^-/)

			if i.index("-tag")

				# the following ARG is actually for -tag, not the main script

				@@opts << i + " " + @@inputs[@@inputs.index(i) + 1]

				@@inputs.delete_at(@@inputs.index(i) + 1)

			else

				@@opts << i

			end

		else

			@@mods << i

		end

	end

	def self.get_opts		

		return @@opts

	end

	def self.get_mods

		return @@mods

	end

end

