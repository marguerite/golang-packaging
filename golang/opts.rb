module Opts

	@@inputs = ['-buildmode=shared','-linkshared','--cflags="a b c"','--with-buildid','-tags','tag','a','b...']

	#@@inputs = ARGV

	@@opts, @@mods = [], []

	def self.get

		# ARGV: -buildmode=shared -linkshared -cflags="a b c" --with-buildid -tags tag d e...
		@@inputs.each do |i|

			# options begin with "-"
			if i.index(/^-/)

				if i.index("-tag")

					# the following ARG is actually for -tag, not the main script

					@@opts << "#{i} #{@@inputs[@@inputs.index(i) + 1]}"

					@@inputs.delete_at(@@inputs.index(i) + 1)

				else

					@@opts << i

				end

			else

				@@mods << i

			end

		end		

		return @@opts, @@mods

	end

end

puts Opts.get
