#!/usr/bin/env ruby

require 'fileutils'
require 'openssl'
require 'openssl-extensions/all'
require 'pathname'

class CertGenerator
  attr_accessor :root_cert, :root_key

  def initialize
    @root_key = OpenSSL::PKey::RSA.new 2048 # the CA's public/private key
    @root_cert = OpenSSL::X509::Certificate.new
    root_cert.version = 2 # cf. RFC 5280 - to make it a "v3" certificate
    root_cert.subject = OpenSSL::X509::Name.parse "/CN=ManageIQ CA"
    root_cert.issuer = root_cert.subject # root CA's are "self-signed"
    root_cert.public_key = root_key.public_key
    root_cert.not_before = Time.now
    root_cert.not_after = root_cert.not_before + 2 * 365 * 24 * 60 * 60 # 2 years
    ef = OpenSSL::X509::ExtensionFactory.new
    ef.subject_certificate = root_cert
    ef.issuer_certificate = root_cert
    root_cert.add_extension(ef.create_extension("basicConstraints","CA:TRUE",true))
    root_cert.add_extension(ef.create_extension("keyUsage","keyCertSign, cRLSign", true))
    root_cert.add_extension(ef.create_extension("subjectKeyIdentifier","hash",false))
    root_cert.add_extension(ef.create_extension("authorityKeyIdentifier","keyid:always",false))
    root_cert.sign(root_key, OpenSSL::Digest::SHA256.new)

    write_files("root", root_cert, root_key)
  end

  def generate_cert(dest, *sans)
    san = sans.collect { |name| "DNS:#{name}"}.join(",")
    key = OpenSSL::PKey::RSA.new 2048
    cert = OpenSSL::X509::Certificate.new
    cert.version = 2
    cert.subject = OpenSSL::X509::Name.parse "/CN=#{dest}"
    cert.issuer = root_cert.subject # root CA is the issuer
    cert.public_key = key.public_key
    cert.not_before = Time.now
    cert.not_after = cert.not_before + 2 * 365 * 24 * 60 * 60 # 2 years
    ef = OpenSSL::X509::ExtensionFactory.new
    ef.subject_certificate = cert
    ef.issuer_certificate = root_cert
    cert.add_extension(ef.create_extension("subjectAltName", san, false)) if san != ""
    cert.add_extension(ef.create_extension("keyUsage","digitalSignature", true))
    cert.add_extension(ef.create_extension("subjectKeyIdentifier","hash",false))
    cert.sign(root_key, OpenSSL::Digest::SHA256.new)

    write_files(dest, cert, key)
  end

  private def certs_path
    @certs_path ||= Pathname.new(__dir__).join("certs").tap { |p| FileUtils.mkdir_p(p.to_s) }
  end

  private def write_files(name, cert, key)
    File.write(certs_path.join("#{name}.crt"), cert)
    File.write(certs_path.join("#{name}.key"), key)
  end
end


application_domain = ARGV[0]
if ARGV.length != 1 || application_domain.length == 0
  puts "Usage: cert_generator your.application.domain.example.com"
  exit 1
end

c = CertGenerator.new
c.generate_cert("httpd")
c.generate_cert("kafka")
c.generate_cert("memcached")
c.generate_cert("postgresql")

c.generate_cert("api", application_domain)
c.generate_cert("remote-console", application_domain)
c.generate_cert("ui", application_domain)
