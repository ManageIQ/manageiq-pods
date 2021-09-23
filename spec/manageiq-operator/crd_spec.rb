require 'awesome_spawn'
describe "Generate CRDs" do
  it "doesn't change any tracked files" do
    unless assert_operator_sdk_valid
      skip("Wrong operator-sdk version installed.  Expected #{expected_operator_sdk_version.inspect}, Found #{operator_sdk_version.inspect}.")
    end

    csv_version = File.read("manageiq-operator/version/version.go").match(/Version\s=\s\"(.+)\"/)[1]

    AwesomeSpawn.run!("operator-sdk generate k8s", :chdir => ROOT.join("manageiq-operator"))
    AwesomeSpawn.run!("operator-sdk generate crds", :chdir => ROOT.join("manageiq-operator"))
    AwesomeSpawn.run!("operator-sdk generate csv --update-crds --csv-version=#{csv_version}", :chdir => ROOT.join("manageiq-operator"))

    expect(AwesomeSpawn.run!("git diff").output).to be_empty, "Files differ after generating CRDs"
  end
end
