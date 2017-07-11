describe "Template Load and Dump" do
  context "remains unchanged" do
    Dir[ROOT.join("templates/**/*.yaml")].each do |file|
      it "for #{file}" do
        require 'yaml'
        content = File.read(file)
        dumped  = YAML.dump(YAML.load(content), :line_width => -1).sub(/^---\n/, "").gsub("\s\n", "\n")

        expect(content).to eq(dumped)
      end
    end
  end
end
