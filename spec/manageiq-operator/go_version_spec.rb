describe "Go Version" do
  it "is the expected version" do
    go_mod = GoMod.new(ROOT.join("manageiq-operator", "go.mod"))
    go_mod.parse

    local_version = Gem::Version.new(go_mod.extract_version(`go version`))

    go_mod.go_version.segments.each_with_index { |s, i| expect(s).to eq(local_version.segments[i]), "expected go version #{go_mod.go_version}, got #{local_version}" }
  end
end
