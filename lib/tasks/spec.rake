require 'rspec/core/rake_task'

RSpec::Core::RakeTask.new(:spec)

desc "Run Go tests"
task :go_test do
  sh "go test -v ./...", :chdir => File.expand_path("../../manageiq-operator", __dir__)
  puts
end

task :default => [:spec, :go_test]
