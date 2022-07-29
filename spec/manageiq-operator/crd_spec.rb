require 'awesome_spawn'
describe "Generate CRDs" do
  it "doesn't change any tracked files" do
    # TODO: Figure out how to make this generate the correct data.
    #       We are running a namespaced operator, but don't want to specify the namespace name in files
    #       since that's up to the person installing the app.
    # AwesomeSpawn.run!("make manifests", :chdir => ROOT.join("manageiq-operator"))

    AwesomeSpawn.run!("make generate", :chdir => ROOT.join("manageiq-operator"))

    expect(AwesomeSpawn.run!("git diff").output).to be_empty, "Files differ after generating CRDs"
  end
end
