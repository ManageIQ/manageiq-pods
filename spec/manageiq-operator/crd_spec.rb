require 'awesome_spawn'
describe "Generate CRDs" do
  it "doesn't change any tracked files" do
    AwesomeSpawn.run!("operator-sdk generate k8s", :chdir => ROOT.join("manageiq-operator"))
    AwesomeSpawn.run!("operator-sdk generate crds", :chdir => ROOT.join("manageiq-operator"))
    AwesomeSpawn.run!("operator-sdk generate csv --update-crds --csv-version=0.0.1", :chdir => ROOT.join("manageiq-operator"))

    expect(AwesomeSpawn.run!("git diff").output).to be_empty, "Files differ after generating CRDs"
  end
end
