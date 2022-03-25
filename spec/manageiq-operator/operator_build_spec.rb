describe "Operator build check" do
  it "successfully builds" do
    require 'awesome_spawn'
    expect { AwesomeSpawn.run!("make build", :chdir => ROOT.join("manageiq-operator")) }.not_to raise_error
  end
end
