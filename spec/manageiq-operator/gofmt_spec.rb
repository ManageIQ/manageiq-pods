describe "Go formatting check" do
  it "has no changes" do
    require 'awesome_spawn'
    result = AwesomeSpawn.run!("gofmt -l manageiq-operator", :chdir => ROOT)
    expect(result.error).to be_empty
    expect(result.output).to be_empty
  end
end
