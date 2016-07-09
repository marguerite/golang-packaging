module CLI

  require 'fileutils'
  require 'timeout'

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

    if $? == 1
      FileUtils.touch "/tmp/failed"
      abort '[ERROR]Go command failed! Please check.'
    end

  end

  def self.run(env={},cmd="")

    puts "GOPATH: #{env}"
    puts "Command: #{cmd}"
    unless RUBY_VERSION.to_f > 1.8
      popen_env(env,cmd) {|f| f.each_line {|l| puts l}}
    else
      IO.popen(env,cmd) {|f| f.each_line {|l| puts l}}
      if $? != 0
        FileUtils.touch "/tmp/failed"
        abort "[ERROR]Go command faild! Please check."
      end
    end

  end

end
