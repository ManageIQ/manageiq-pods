require 'awesome_spawn'
describe "Generate CRDs" do
  it "doesn't change any tracked files" do
    AwesomeSpawn.run!("make manifests", :chdir => ROOT.join("manageiq-operator"))

    AwesomeSpawn.run!("make generate", :chdir => ROOT.join("manageiq-operator"))

    diff_output = AwesomeSpawn.run!("git diff manageiq-operator/").output
    puts diff_output unless diff_output.empty?
    expect(diff_output).to be_empty, "Files differ after generating CRDs"
  end
end
