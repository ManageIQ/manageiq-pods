describe "Operator build check" do
  it "successfully builds" do
    require 'awesome_spawn'
    expect { AwesomeSpawn.run!("go build -o build/_output/bin/manageiq-operator ./cmd/manager", :chdir => ROOT.join("manageiq-operator")) }.not_to raise_error
  end
end
