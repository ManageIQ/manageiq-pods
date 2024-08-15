describe "Go Version" do
  it "matches go.mod" do
    require 'awesome_spawn'
    dockerfile_version = File.read(ROOT.join("manageiq-operator", "Dockerfile")).match(/^FROM.+golang:(\d+\.\d+).+/)[1]
    mod_version        = File.read(ROOT.join("manageiq-operator", "go.mod")).match(/^go\s(\d+\.\d+)/)[1]
    running_version    = AwesomeSpawn.run!("go version", :chdir => ROOT.join("manageiq-operator")).output.match(/.*\sgo(\d+\.\d+).*/)[1]

    expect(dockerfile_version).to eq(mod_version)
    expect(running_version).to    eq(mod_version)
  end
end
