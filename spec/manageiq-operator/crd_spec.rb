require 'awesome_spawn'
describe "Generate CRDs" do
  it "doesn't change any tracked files" do
    AwesomeSpawn.run!("make manifests", :chdir => ROOT.join("manageiq-operator"))

    AwesomeSpawn.run!("make generate", :chdir => ROOT.join("manageiq-operator"))

    expect(AwesomeSpawn.run!("git diff").output).to be_empty, "Files differ after generating CRDs"
  end
end
