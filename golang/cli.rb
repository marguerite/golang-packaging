module CLI

	def self.run(env={},cmd="")

		# newer version of ruby doesn't have return value for popen3
		# while older version of ruby can't pass array to popen
		unless RUBY_VERSION.to_f > 1.8
			require 'open3'
			Open3.popen3(env,cmd) {|s1,s2,s3| s2.each_line {|l| puts l}}
		else
			IO.popen(env,cmd) {|f| f.each_line {|l| puts l}}
		end

		if $? == 0
			File.open("/tmp/exitstatus.txt","w:UTF-8") {|f| f.puts(0)}
		else
			File.open("/tmp/exitstatus.txt","w:UTF-8") {|f| f.puts(1)}
			abort "[ERROR]Go command failed! Please check."
		end

	end

end

