module CLI

  require 'timeout'

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

  # popen in 1.8 doesn't support env hash
  def popen_env(hash, cmd)
    # set ENV separately
    hash.each {|k,v| ENV[k] = v}

    # some commands need time, an immediate close
    # will get a wrong status, so wait them with
    # timeout 30s
    # set timeout 300s, because go install takes
    # lots of time sometimes
    begin
      Timeout.timeout(300) do		
        @pipe = IO.popen(cmd)
        Process.wait(@pipe.pid)
      end
    rescue Timeout::Error
      Process.kill(9,@pipe.pid)
      # collect status
      Process.wait(@pipe.pid) 
    end

    write_status($?)

  end

  def self.run(env={},cmd="")

    puts "GOPATH: #{env}"
    puts "Command: #{cmd}"
    unless RUBY_VERSION.to_f > 1.8
      popen_env(env,cmd) {|f| f.each_line {|l| puts l}}
    else
      IO.popen(env,cmd) {|f| f.each_line {|l| puts l}}
      write_status($?)
    end

  end

end

