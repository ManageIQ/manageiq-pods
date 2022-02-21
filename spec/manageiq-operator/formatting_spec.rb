require 'awesome_spawn'

describe "Go formatting check" do
  it "has no source changes" do
    result = AwesomeSpawn.run!("gofmt -l manageiq-operator", :chdir => ROOT)
    expect(result.error).to be_empty
    expect(result.output.split("\n") - ["manageiq-operator/pkg/apis/manageiq/v1alpha1/zz_generated.deepcopy.go"]).to be_empty
  end

  it "has no mod/sum changes" do
    result = AwesomeSpawn.run!("go mod tidy", :chdir => ROOT.join("manageiq-operator"))
    expect(result.output).to be_empty

    expect(
      AwesomeSpawn.run!("git diff manageiq-operator/go.mod manageiq-operator/go.sum").output
    ).to be_empty, "Files differ after running go mod tidy"
  end
end
