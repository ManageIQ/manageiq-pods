class GoMod
  attr_reader :path, :go_version, :requires

  def initialize(path)
    require 'pathname'
    @path     = Pathname.new(path)
    @requires = {}
  end

  def parse
    return if @parsed
    parse_block = nil

    path.read.each_line do |line|
      if line.start_with?("go ")
        parse_go_version(line)
        next
      elsif line.start_with?("require (")
        parse_block = :require
      elsif line.start_with?(")")
        parse_block = nil
      elsif parse_block
        send("add_#{parse_block}", line)
      end
    end

    @parsed = true
  end

  private def parse_go_version(line)
    @go_version = Gem::Version.new(line.split.last)
  end

  private def add_require(line)
    package, version   = line.split
    @requires[package] = Gem::Version.new(extract_version(version))
  end

  def extract_version(string)
    string.match(/(\d+(\.\d+(\.\d+)?)?).*/)[1]
  end
end
