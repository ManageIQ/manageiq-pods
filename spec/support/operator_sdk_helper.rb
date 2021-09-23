def expected_operator_sdk_version
  @expected_operator_sdk_version ||= ROOT.join("manageiq-operator", "go.mod").readlines.grep(/operator-sdk/).first.strip.split(" ").last
end

def operator_sdk_version
  return @operator_sdk_version if defined?(@operator_sdk_version)

  version_str = AwesomeSpawn.run!("operator-sdk version").output
  @operator_sdk_version = version_str.match(/operator-sdk version: "([^"]+)"/)[1]
rescue
  @operator_sdk_version = nil
end

def assert_operator_sdk_valid
  (operator_sdk_version == expected_operator_sdk_version).tap do |valid|
    raise "Invalid operator-sdk version: #{operator_sdk_version}" if ENV["CI"] && !valid
  end
end
