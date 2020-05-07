package miqtools

import (
	"fmt"
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
)

// auth-configuration.conf
func httpdAuthConfigurationConf() string {
	return `
# External Authentication Configuration File
#
# For details on usage please see https://github.com/ManageIQ/manageiq-pods/blob/master/README.md#configuring-external-authentication
`
}

// application.conf
func httpdApplicationConf() string {
	return `
Listen 8080
# Timeout: The number of seconds before receives and sends time out.
Timeout 120

RewriteEngine On
Options SymLinksIfOwnerMatch

<VirtualHost *:8080>
  KeepAlive on
  # Without ServerName mod_auth_mellon compares against http:// and not https:// from the IdP
  ServerName https://%{REQUEST_HOST}

  ProxyPreserveHost on

  RewriteCond %{REQUEST_URI}     ^/ws        [NC]
  RewriteCond %{HTTP:UPGRADE}    ^websocket$ [NC]
  RewriteCond %{HTTP:CONNECTION} ^Upgrade$   [NC]
  RewriteRule .* ws://websocket:3000%{REQUEST_URI}  [P,QSA,L]
  ProxyPassReverse /ws ws://websocket:3000/ws

  # For httpd, some ErrorDocuments must by served by the httpd pod
  RewriteCond %{REQUEST_URI} !^/proxy_pages

  # For SAML /saml2 is only served by mod_auth_mellon in the httpd pod
  RewriteCond %{REQUEST_URI} !^/saml2

  # For OpenID-Connect /openid-connect is only served by mod_auth_openidc
  RewriteCond %{REQUEST_URI} !^/openid-connect

  RewriteRule ^/ http://ui:3000%{REQUEST_URI} [P,QSA,L]
  ProxyPassReverse / http://ui:3000/

  ProxyPass /api http://web-service:3000/api
  ProxyPassReverse /api http://web-service:3000/api

  # Ensures httpd stdout/stderr are seen by 'docker logs'.
  ErrorLog  "| /usr/bin/tee /proc/1/fd/2 /var/log/httpd/error_log"
  CustomLog "| /usr/bin/tee /proc/1/fd/1 /var/log/httpd/access_log" common
</VirtualHost>
`
}

// authentication.conf
func httpdAuthenticationConf(spec *miqv1alpha1.ManageIQSpec) string {
	switch spec.HttpdAuthenticationType {
	case "openid-connect":
		return httpdOIDCAuthConf(spec.OIDCProviderURL, spec.ApplicationDomain)
	case "external":
		return httpdExternalAuthConf()
	case "active-directory":
		return httpdADAuthConf()
	case "saml":
		return httpdSAMLAuthConf()
	default:
		return ""
	}
}

func httpdExternalAuthConf() string {
	s := `
%s

<Location /dashboard/kerberos_authenticate>
  AuthType           GSSAPI
  AuthName           "GSSAPI Single Sign On Login"
  GssapiCredStore    keytab:/etc/http.keytab
  GssapiLocalName    on
  Require            pam-account httpd-auth

  ErrorDocument 401 /proxy_pages/invalid_sso_credentials.js
</Location>

%s
%s
%s
%s
`
	return fmt.Sprintf(s, httpdAuthLoadModulesConf(), httpdAuthLoginFormConf(), httpdAuthApplicationAPIConf(), httpdAuthLookupUserDetailsConf(), httpdAuthRemoteUserConf())
}

func httpdADAuthConf() string {
	s := `
%s

<Location /dashboard/kerberos_authenticate>
  AuthType           GSSAPI
  AuthName           "GSSAPI Single Sign On Login"
  GssapiCredStore    keytab:/etc/krb5.keytab
  GssapiLocalName    on
  Require            pam-account httpd-auth

  ErrorDocument 401 /proxy_pages/invalid_sso_credentials.js
</Location>

%s
%s
%s
%s
`
	return fmt.Sprintf(s, httpdAuthLoadModulesConf(), httpdAuthLoginFormConf(), httpdAuthApplicationAPIConf(), httpdAuthLookupUserDetailsConf(), httpdAuthRemoteUserConf())
}

func httpdSAMLAuthConf() string {
	s := `
LoadModule auth_mellon_module modules/mod_auth_mellon.so

<Location />
  MellonEnable               "info"

  MellonIdPMetadataFile      "/etc/httpd/saml2/idp-metadata.xml"

  MellonSPPrivateKeyFile     "/etc/httpd/saml2/sp-key.key"
  MellonSPCertFile           "/etc/httpd/saml2/sp-cert.cert"
  MellonSPMetadataFile       "/etc/httpd/saml2/sp-metadata.xml"

  MellonVariable             "sp-cookie"
  MellonSecureCookie         On
  MellonCookiePath           "/"

  MellonIdP                  "IDP"

  MellonEndpointPath         "/saml2"

  MellonUser                 username
  MellonMergeEnvVars         On

  MellonSetEnvNoPrefix       "REMOTE_USER"            username
  MellonSetEnvNoPrefix       "REMOTE_USER_EMAIL"      email
  MellonSetEnvNoPrefix       "REMOTE_USER_FIRSTNAME"  firstname
  MellonSetEnvNoPrefix       "REMOTE_USER_LASTNAME"   lastname
  MellonSetEnvNoPrefix       "REMOTE_USER_FULLNAME"   fullname
  MellonSetEnvNoPrefix       "REMOTE_USER_GROUPS"     groups
</Location>

<Location /saml_login>
  AuthType                   "Mellon"
  MellonEnable               "auth"
  Require                    valid-user
</Location>

%s
`
	return fmt.Sprintf(s, httpdAuthRemoteUserConf())
}

func httpdOIDCAuthConf(providerURL, applicationDomain string) string {
	// If these are not provided, we should assume that the user provided a full config
	// in a secret, so include the directory for that secret here
	if providerURL == "" || applicationDomain == "" {
		return "Include user-conf.d/*.conf"
	}

	s := `
LoadModule auth_openidc_module modules/mod_auth_openidc.so

OIDCProviderMetadataURL      %s
OIDCClientID                 ${HTTPD_AUTH_OIDC_CLIENT_ID}
OIDCClientSecret             ${HTTPD_AUTH_OIDC_CLIENT_SECRET}

OIDCRedirectURI              "https://%s/oidc_login/redirect_uri"
OIDCOAuthRemoteUserClaim     username

OIDCCryptoPassphrase         sp-secret

<Location /oidc_login>
  AuthType                   openid-connect
  Require                    valid-user
</Location>

RequestHeader unset X_REMOTE_USER

RequestHeader set X_REMOTE_USER           %%{OIDC_CLAIM_PREFERRED_USERNAME}e env=OIDC_CLAIM_PREFERRED_USERNAME
RequestHeader set X_EXTERNAL_AUTH_ERROR   %%{EXTERNAL_AUTH_ERROR}e           env=EXTERNAL_AUTH_ERROR
RequestHeader set X_REMOTE_USER_EMAIL     %%{OIDC_CLAIM_EMAIL}e              env=OIDC_CLAIM_EMAIL
RequestHeader set X_REMOTE_USER_FIRSTNAME %%{OIDC_CLAIM_GIVEN_NAME}e         env=OIDC_CLAIM_GIVEN_NAME
RequestHeader set X_REMOTE_USER_LASTNAME  %%{OIDC_CLAIM_FAMILY_NAME}e        env=OIDC_CLAIM_FAMILY_NAME
RequestHeader set X_REMOTE_USER_FULLNAME  %%{OIDC_CLAIM_NAME}e               env=OIDC_CLAIM_NAME
RequestHeader set X_REMOTE_USER_GROUPS    %%{OIDC_CLAIM_GROUPS}e             env=OIDC_CLAIM_GROUPS
RequestHeader set X_REMOTE_USER_DOMAIN    %%{OIDC_CLAIM_DOMAIN}e             env=OIDC_CLAIM_DOMAIN
`
	return fmt.Sprintf(s, providerURL, applicationDomain)
}

func httpdAuthLoadModulesConf() string {
	return `
LoadModule authnz_pam_module            modules/mod_authnz_pam.so
LoadModule intercept_form_submit_module modules/mod_intercept_form_submit.so
LoadModule lookup_identity_module       modules/mod_lookup_identity.so
LoadModule auth_kerb_module             modules/mod_auth_kerb.so
`
}

func httpdAuthLoginFormConf() string {
	return `
<Location /dashboard/external_authenticate>
	InterceptFormPAMService    httpd-auth
	InterceptFormLogin         user_name
	InterceptFormPassword      user_password
	InterceptFormLoginSkip     admin
	InterceptFormClearRemoteUserForSkipped on
</Location>
`
}

func httpdAuthApplicationAPIConf() string {
	return `
<LocationMatch ^/api(?!\/(v[\d\.]+\/)?product_info$)>
  SetEnvIf Authorization '^Basic +YWRtaW46' let_admin_in
  SetEnvIf X-Auth-Token  '^.+$'             let_api_token_in
  SetEnvIf X-MIQ-Token   '^.+$'             let_sys_token_in
  SetEnvIf X-CSRF-Token  '^.+$'             let_csrf_token_in

  AuthType Basic
  AuthName "External Authentication (httpd) for API"
  AuthBasicProvider PAM

  AuthPAMService httpd-auth
  Require        valid-user
  Order          Allow,Deny
  Allow from env=let_admin_in
  Allow from env=let_api_token_in
  Allow from env=let_sys_token_in
  Allow from env=let_csrf_token_in
  Satisfy Any

  LookupUserAttr mail        REMOTE_USER_EMAIL
  LookupUserAttr givenname   REMOTE_USER_FIRSTNAME
  LookupUserAttr sn          REMOTE_USER_LASTNAME
  LookupUserAttr displayname REMOTE_USER_FULLNAME
  LookupUserAttr domainname  REMOTE_USER_DOMAIN

  LookupUserGroups           REMOTE_USER_GROUPS ":"
  LookupDbusTimeout          5000

</LocationMatch>
`
}

func httpdAuthLookupUserDetailsConf() string {
	return `
<LocationMatch ^/dashboard/external_authenticate$|^/dashboard/kerberos_authenticate$|^/api>
	LookupUserAttr mail        REMOTE_USER_EMAIL
	LookupUserAttr givenname   REMOTE_USER_FIRSTNAME
	LookupUserAttr sn          REMOTE_USER_LASTNAME
	LookupUserAttr displayname REMOTE_USER_FULLNAME
	LookupUserAttr domainname  REMOTE_USER_DOMAIN

	LookupUserGroups           REMOTE_USER_GROUPS ":"
	LookupDbusTimeout          5000
</LocationMatch>
`
}

func httpdAuthRemoteUserConf() string {
	return `
RequestHeader unset X_REMOTE_USER

RequestHeader set X_REMOTE_USER           %{REMOTE_USER}e           env=REMOTE_USER
RequestHeader set X_EXTERNAL_AUTH_ERROR   %{EXTERNAL_AUTH_ERROR}e   env=EXTERNAL_AUTH_ERROR
RequestHeader set X_REMOTE_USER_EMAIL     %{REMOTE_USER_EMAIL}e     env=REMOTE_USER_EMAIL
RequestHeader set X_REMOTE_USER_FIRSTNAME %{REMOTE_USER_FIRSTNAME}e env=REMOTE_USER_FIRSTNAME
RequestHeader set X_REMOTE_USER_LASTNAME  %{REMOTE_USER_LASTNAME}e  env=REMOTE_USER_LASTNAME
RequestHeader set X_REMOTE_USER_FULLNAME  %{REMOTE_USER_FULLNAME}e  env=REMOTE_USER_FULLNAME
RequestHeader set X_REMOTE_USER_GROUPS    %{REMOTE_USER_GROUPS}e    env=REMOTE_USER_GROUPS
RequestHeader set X_REMOTE_USER_DOMAIN    %{REMOTE_USER_DOMAIN}e    env=REMOTE_USER_DOMAIN
`
}
