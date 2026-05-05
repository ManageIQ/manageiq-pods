require 'active_support/core_ext/module/delegation'
require 'pathname'

require_relative './manageiq_user_mixin'

module ManageIQ
  module ApplianceConsole
    class MessageConfiguration
      include ManageIQ::ApplianceConsole::ManageiqUserMixin
      include ManageIQ::ApplianceConsole::Prompts

      attr_reader :message_keystore_username, :message_keystore_password,
                  :message_server_host, :message_server_port,
                  :miq_config_dir_path, :config_dir_path, :sample_config_dir_path,
                  :client_properties_path,
                  :keystore_dir_path, :truststore_path, :keystore_path,
                  :messaging_yaml_sample_path, :messaging_yaml_path,
                  :ca_cert_path

      BASE_DIR                          = "/var/lib/kafka".freeze
      LOGS_DIR                          = "#{BASE_DIR}/logs".freeze
      CONFIG_DIR                        = "#{BASE_DIR}/config".freeze
      SAMPLE_CONFIG_DIR                 = "#{BASE_DIR}/config-sample".freeze
      MIQ_CONFIG_DIR                    = ManageIQ::ApplianceConsole::RAILS_ROOT.join("config").freeze

      def self.available?
        File.exist?("#{BASE_DIR}/bin/kafka-run-class.sh")
      end

      def self.configured?
        MessageServerConfiguration.configured? || MessageClientConfiguration.configured?
      end

      def initialize(options = {})
        @message_server_port        = options[:message_server_port] || 9093
        @message_keystore_username  = options[:message_keystore_username] || "admin"
        @message_keystore_password  = options[:message_keystore_password]

        @miq_config_dir_path        = Pathname.new(MIQ_CONFIG_DIR)
        @config_dir_path            = Pathname.new(CONFIG_DIR)
        @sample_config_dir_path     = Pathname.new(SAMPLE_CONFIG_DIR)

        @client_properties_path     = config_dir_path.join("client.properties")
        @keystore_dir_path          = config_dir_path.join("keystore")
        @truststore_path            = keystore_dir_path.join("truststore.jks")
        @keystore_path              = keystore_dir_path.join("keystore.jks")

        @messaging_yaml_sample_path = miq_config_dir_path.join("messaging.kafka.yml")
        @messaging_yaml_path        = miq_config_dir_path.join("messaging.yml")
        @ca_cert_path               = keystore_dir_path.join("ca-cert")
      end

      def already_configured?
        installed_file_found = false
        installed_files.each do |f|
          if File.exist?(f)
            installed_file_found = true
            say("Installed file #{f} found.")
          end
        end
        installed_file_found
      end

      def ask_questions
        return false unless valid_environment?

        ask_for_parameters
        show_parameters
        return false unless agree("\nProceed? (Y/N): ")

        return false unless host_resolvable?(message_server_host) && host_reachable?(message_server_host, "Message Server Host:")

        true
      end

      def create_client_properties
        say(__method__.to_s.tr("_", " ").titleize)

        return if file_found?(client_properties_path)

        algorithm = message_server_host.ipaddress? ? "" : "HTTPS"
        protocol = secure? ? "SASL_SSL" : "PLAINTEXT"
        content = secure? ? secure_client_properties_content(algorithm, protocol) : unsecure_client_properties_content(algorithm, protocol)

        File.write(client_properties_path, content)
      end

      def secure_client_properties_content(algorithm, protocol)
        secure_content = <<~CLIENT_PROPERTIES
          ssl.truststore.location=#{truststore_path}
          ssl.truststore.password=#{message_keystore_password}
        CLIENT_PROPERTIES

        unsecure_client_properties_content(algorithm, protocol) + secure_content
      end

      def unsecure_client_properties_content(algorithm, protocol)
        <<~CLIENT_PROPERTIES
          ssl.endpoint.identification.algorithm=#{algorithm}

          sasl.mechanism=PLAIN
          security.protocol=#{protocol}
          sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required \\
            username=#{message_keystore_username} \\
            password=#{message_keystore_password} ;
        CLIENT_PROPERTIES
      end

      def configure_messaging_yaml
        say(__method__.to_s.tr("_", " ").titleize)

        return if file_found?(messaging_yaml_path)

        data = File.read(messaging_yaml_sample_path)
        messaging_yaml =
          if YAML.respond_to?(:safe_load)
            YAML.safe_load(data, :aliases => true)
          else
            YAML.load(data) # rubocop:disable Security/YAMLLoad
          end

        messaging_yaml["production"]["host"]      = message_server_host
        messaging_yaml["production"]["port"]      = message_server_port
        messaging_yaml["production"]["username"]  = message_keystore_username
        messaging_yaml["production"]["password"]  = ManageIQ::Password.try_encrypt(message_keystore_password)

        if secure?
          messaging_yaml["production"]["ssl"]     = true
          messaging_yaml["production"]["ca_file"] = ca_cert_path.to_path
        else
          messaging_yaml["production"]["ssl"] = false
        end

        File.open(messaging_yaml_path, "w") do |f|
          f.write(messaging_yaml.to_yaml)
          f.chown(manageiq_uid, manageiq_gid)
        end
      end

      def remove_installed_files
        say(__method__.to_s.tr("_", " ").titleize)

        installed_files.each { |f| FileUtils.rm_rf(f) }
      end

      def valid_environment?
        if already_configured?
          unconfigure if agree("\nAlready configured on this Appliance, Un-Configure first? (Y/N): ")
          return false unless agree("\nProceed with Configuration? (Y/N): ")
        end
        true
      end

      def file_found?(path)
        return false unless File.exist?(path)

        say("\tWARNING: #{path} already exists. Taking no action.")
        true
      end

      def files_found?(path_list)
        return false unless path_list.all? { |path| File.exist?(path) }

        path_list.each { |path| file_found?(path) }
        true
      end

      def file_contains?(path, content)
        return false unless File.exist?(path)

        content.split("\n").each do |l|
          l.gsub!("/", "\\/")
          l.gsub!(/password=.*$/, "password=") # Remove the password as it can have special characters that grep can not match.
          return false unless File.foreach(path).grep(/#{l}/).any?
        end

        say("Content already exists in #{path}. Taking no action.")
        true
      end

      def host_reachable?(host, what)
        require 'net/ping'
        say("Checking connectivity to #{host} ... ")
        unless Net::Ping::External.new(host).ping
          say("Failed.\nCould not connect to #{host},")
          say("the #{what} must be reachable by name.")
          return false
        end
        say("Succeeded.")
        true
      end

      def host_resolvable?(host)
        require 'ipaddr'
        require 'resolv'

        say("Checking if #{host} is resolvable ... ")
        begin
          ip_address = Resolv.getaddress(host)

          if IPAddr.new("127.0.0.1/8").include?(ip_address) || IPAddr.new("::1/128").include?(ip_address)
            say("Failed.\nThe hostname must not resolve to a link-local address")

            return false
          end
        rescue Resolv::ResolvError => e
          say("Failed.\nHostname #{host} is not resolvable: #{e.message}")

          return false
        end

        say("Succeeded.")
        true
      end

      def unconfigure
        remove_installed_files
      end

      def secure?
        message_server_port == 9_093
      end
    end
  end
end
