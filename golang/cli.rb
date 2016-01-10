module CLI

	def self.run(command="")

		# echo the command we run to the buildlog
		puts command

		IO.popen(command) {|f| f.each_line {|l| puts l}}

		if $? == 0
			File.open("/tmp/exitstatus.txt","w:UTF-8") {|f| f.puts(0)}
		else
			File.open("/tmp/exitstatus.txt","w:UTF-8") {|f| f.puts(1)}
			abort "[ERROR]Go command failed! Please check."
		end

	end

end

