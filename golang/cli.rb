module CLI

	def write_status(status)
		file = "/tmp/exitstatus.txt"
		mode = "w:UTF-8"
		if status == 0
			File.open(file,mode) {|f| f.puts(0)}
		else
			File.open(file,mode) {|f| f.puts(1)}
			abort "[ERROR]Go command failed! Please check."
		end
	end

	def self.run(env={},cmd="")
		unless RUBY_VERSION.to_f > 1.8
			# popen in 1.8 doesn't support env hash
			def popen_env(hash, cmd)
				hash.each do |k,v|
					ENV[k] = v
				end
				io = IO.popen(cmd)
				io.close
				write_status($?)
			end
			popen_env(env,cmd) {|f| f.each_line {|l| puts l}}
		else
			IO.popen(env,cmd) {|f| f.each_line {|l| puts l}}
			write_status($?)
		end
	end

end

