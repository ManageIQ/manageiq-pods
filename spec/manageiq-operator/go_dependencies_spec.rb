describe "Go dependencies" do
  let(:go_mod) { GoMod.new(ROOT.join("manageiq-operator", "go.mod")).tap(&:parse) }

  def version_ok?(expected, given)
    expected.segments.each_with_index { |s, i| expect(s).to eq(given.segments[i]), "expected: #{expected}, given: #{given}" }
  end

  it "'go' is the expected version" do
    installed_go_version = Gem::Version.new(go_mod.extract_version(`go version`))

    version_ok?(go_mod.go_version, installed_go_version)
  end

  it "'operator-sdk' is the expected version" do
    installed_operator_sdk_version = Gem::Version.new(go_mod.extract_version(`operator-sdk version`))

    version_ok?(go_mod.requires["github.com/operator-framework/operator-sdk"], installed_operator_sdk_version)
  end
end
