package miqtools

import (
	"fmt"
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
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
func httpdApplicationConf(applicationDomain string, uiHttpProtocol string, uiWebSocketProtocol string, apiHttpProtocol string) string {
	s := `
Listen 8080

# Timeout: The number of seconds before receives and sends time out.
Timeout 120
ServerSignature Off
ServerTokens Prod

RewriteEngine On
Options SymLinksIfOwnerMatch

<VirtualHost *:8080>
  IncludeOptional conf.d/ssl_config
  IncludeOptional conf.d/ssl_proxy_config

  KeepAlive on
  # Without ServerName mod_auth_mellon compares against http:// and not https:// from the IdP
  ServerName https://%%{REQUEST_HOST}

  ProxyPreserveHost on
  RequestHeader set Host %[1]s
  RequestHeader set X-Forwarded-Host %[1]s

  RewriteCond %%{REQUEST_URI}     ^/ws/notifications [NC]
  RewriteCond %%{HTTP:UPGRADE}    ^websocket$ [NC]
  RewriteCond %%{HTTP:CONNECTION} ^Upgrade$   [NC]
  RewriteRule .* %[2]s://ui:3000%%{REQUEST_URI}  [P,QSA,L]
  ProxyPassReverse /ws/notifications %[2]s://ui:3000/ws/notifications

  RewriteCond %%{REQUEST_URI} !^/api

  # For httpd, some ErrorDocuments must by served by the httpd pod
  RewriteCond %%{REQUEST_URI} !^/proxy_pages

  # For SAML /saml2 is only served by mod_auth_mellon in the httpd pod
  RewriteCond %%{REQUEST_URI} !^/saml2

  # For OpenID-Connect /openid-connect is only served by mod_auth_openidc
  RewriteCond %%{REQUEST_URI} !^/openid-connect

  RewriteRule ^/ %[3]s://ui:3000%%{REQUEST_URI} [P,QSA,L]
  ProxyPassReverse / %[3]s://ui:3000/

  ProxyPass /api %[4]s://web-service:3000/api
  ProxyPassReverse /api %[4]s://web-service:3000/api

  RewriteCond %%{REQUEST_URI}     ^/ws/console [NC]
  RewriteCond %%{HTTP:UPGRADE}    ^websocket$  [NC]
  RewriteCond %%{HTTP:CONNECTION} ^Upgrade$    [NC]
  RewriteRule .* ws://remote-console:3000%%{REQUEST_URI}  [P,QSA,L]
  ProxyPassReverse /ws/console ws://remote-console:3000/ws/console

  # Ensures httpd stdout/stderr are seen by 'docker logs'.
  ErrorLog  "/dev/stderr"
  CustomLog "/dev/stdout" common
</VirtualHost>
`
	return fmt.Sprintf(s, applicationDomain, uiWebSocketProtocol, uiHttpProtocol, apiHttpProtocol)
}

// authentication.conf
func httpdAuthenticationConf(spec *miqv1alpha1.ManageIQSpec) string {
	switch spec.HttpdAuthenticationType {
	case "openid-connect":
		return httpdOIDCAuthConf(spec)
	case "external":
		return httpdExternalAuthConf(*spec.EnableApplicationLocalLogin)
	case "active-directory":
		return httpdADAuthConf(*spec.EnableApplicationLocalLogin)
	case "saml":
		return httpdSAMLAuthConf()
	default:
		return ""
	}
}

func httpdExternalAuthConf(enableLocalLogin bool) string {
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
	apiExtraConfig := `
  AuthBasicProvider PAM
  AuthPAMService httpd-auth

  LookupUserAttr mail        REMOTE_USER_EMAIL
  LookupUserAttr givenname   REMOTE_USER_FIRSTNAME
  LookupUserAttr sn          REMOTE_USER_LASTNAME
  LookupUserAttr displayname REMOTE_USER_FULLNAME
  LookupUserAttr domainname  REMOTE_USER_DOMAIN

  LookupUserGroups           REMOTE_USER_GROUPS ":"
  LookupDbusTimeout          5000
`
	return fmt.Sprintf(
		s,
		httpdAuthLoadModulesConf(),
		httpdAuthLoginFormConf(),
		httpdAuthApplicationAPIConf("Basic", "\"External Authentication (httpd) for API\"", apiExtraConfig, enableLocalLogin),
		httpdAuthLookupUserDetailsConf(),
		httpdAuthRemoteUserConf(":"),
	)
}

func httpdADAuthConf(enableLocalLogin bool) string {
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
	apiExtraConfig := `
  AuthBasicProvider PAM
  AuthPAMService httpd-auth

  LookupUserAttr mail        REMOTE_USER_EMAIL
  LookupUserAttr givenname   REMOTE_USER_FIRSTNAME
  LookupUserAttr sn          REMOTE_USER_LASTNAME
  LookupUserAttr displayname REMOTE_USER_FULLNAME
  LookupUserAttr domainname  REMOTE_USER_DOMAIN

  LookupUserGroups           REMOTE_USER_GROUPS ":"
  LookupDbusTimeout          5000
`
	return fmt.Sprintf(
		s,
		httpdAuthLoadModulesConf(),
		httpdAuthLoginFormConf(),
		httpdAuthApplicationAPIConf("Basic", "\"External Authentication (httpd) for API\"", apiExtraConfig, enableLocalLogin),
		httpdAuthLookupUserDetailsConf(),
		httpdAuthRemoteUserConf(":"),
	)
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
  MellonMergeEnvVars         On ";"

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
	return fmt.Sprintf(s, httpdAuthRemoteUserConf(";"))
}

func httpdOIDCAuthConf(spec *miqv1alpha1.ManageIQSpec) string {
	providerURL := spec.OIDCProviderURL
	introspectionURL := spec.OIDCOAuthIntrospectionURL
	applicationDomain := spec.ApplicationDomain

	// If these are not provided, we should assume that the user provided a full config
	// in a secret, so include the directory for that secret here
	if providerURL == "" || introspectionURL == "" || applicationDomain == "" {
		return "Include user-conf.d/*.conf"
	}

	disableValidation := ""
	if spec.OIDCCACertSecret == "" {
		disableValidation = "OIDCSSLValidateServer Off\nOIDCOAuthSSLValidateServer Off"
	}

	s := `
LoadModule auth_openidc_module modules/mod_auth_openidc.so
ServerName https://%s
LogLevel   warn

OIDCProviderMetadataURL      %s
OIDCClientID                 ${HTTPD_AUTH_OIDC_CLIENT_ID}
OIDCClientSecret             ${HTTPD_AUTH_OIDC_CLIENT_SECRET}
OIDCRedirectURI              "https://%s/oidc_login/redirect_uri"
OIDCCryptoPassphrase         sp-secret
OIDCOAuthRemoteUserClaim     username
OIDCCacheShmEntrySizeMax     65536

OIDCOAuthClientID                  ${HTTPD_AUTH_OIDC_CLIENT_ID}
OIDCOAuthClientSecret              ${HTTPD_AUTH_OIDC_CLIENT_SECRET}
OIDCOAuthIntrospectionEndpoint     %s
OIDCOAuthIntrospectionEndpointAuth client_secret_post
OIDCCookieSameSite                 On

%s

<Location /oidc_login>
  AuthType                   openid-connect
  Require                    valid-user
  FileETag                   None
  Header Set Cache-Control   "max-age=0, no-store, no-cache, must-revalidate"
  Header Set Pragma          "no-cache"
  Header Unset ETag
</Location>

<Location /ui/service/oidc_login>
  AuthType                   openid-connect
  Require                    valid-user
  FileETag                   None
  Header Set Cache-Control   "max-age=0, no-store, no-cache, must-revalidate"
  Header Set Pragma          "no-cache"
  Header Set Set-Cookie      "miq_oidc_access_token=%%{OIDC_access_token}e; Max-Age=10; Path=/ui/service"
  Header Unset ETag
</Location>
%s

RequestHeader unset X-REMOTE-USER
RequestHeader unset X-REMOTE_USER
RequestHeader unset X_REMOTE-USER
RequestHeader unset X_REMOTE_USER

RequestHeader set X_REMOTE_USER                 %%{OIDC_CLAIM_PREFERRED_USERNAME}e env=OIDC_CLAIM_PREFERRED_USERNAME
RequestHeader set X_EXTERNAL_AUTH_ERROR         %%{EXTERNAL_AUTH_ERROR}e           env=EXTERNAL_AUTH_ERROR
RequestHeader set X_REMOTE_USER_EMAIL           %%{OIDC_CLAIM_EMAIL}e              env=OIDC_CLAIM_EMAIL
RequestHeader set X_REMOTE_USER_FIRSTNAME       %%{OIDC_CLAIM_GIVEN_NAME}e         env=OIDC_CLAIM_GIVEN_NAME
RequestHeader set X_REMOTE_USER_LASTNAME        %%{OIDC_CLAIM_FAMILY_NAME}e        env=OIDC_CLAIM_FAMILY_NAME
RequestHeader set X_REMOTE_USER_FULLNAME        %%{OIDC_CLAIM_NAME}e               env=OIDC_CLAIM_NAME
RequestHeader set X_REMOTE_USER_GROUPS          %%{OIDC_CLAIM_GROUPS}e             env=OIDC_CLAIM_GROUPS
RequestHeader set X_REMOTE_USER_GROUP_DELIMITER ","
RequestHeader set X_REMOTE_USER_DOMAIN          %%{OIDC_CLAIM_DOMAIN}e             env=OIDC_CLAIM_DOMAIN
`
	return fmt.Sprintf(
		s,
		applicationDomain,
		providerURL,
		applicationDomain,
		introspectionURL,
		disableValidation,
		httpdAuthApplicationAPIConf("oauth20", "\"External Authentication (oauth20) for API\"", "", *spec.EnableApplicationLocalLogin),
	)
}

func httpdAuthLoadModulesConf() string {
	return `
LoadModule authnz_pam_module            modules/mod_authnz_pam.so
LoadModule intercept_form_submit_module modules/mod_intercept_form_submit.so
LoadModule lookup_identity_module       modules/mod_lookup_identity.so
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

func httpdAuthApplicationAPIConf(authType, authName, extraConfig string, enableLocalLogin bool) string {
	s := `
<LocationMatch ^/api(?!\/(v[\d\.]+\/)?product_info$)>
  SetEnvIf X-Auth-Token  '^.+$'             let_api_token_in
  SetEnvIf X-MIQ-Token   '^.+$'             let_sys_token_in
  SetEnvIf X-CSRF-Token  '^.+$'             let_csrf_token_in

  AuthType %s
  AuthName %s
  Require        valid-user
  Order          Allow,Deny
  Allow from env=let_api_token_in
  Allow from env=let_sys_token_in
  Allow from env=let_csrf_token_in
  Satisfy Any
  %s

  %s
</LocationMatch>
`
	letAdminIn := ""
	if enableLocalLogin {
		letAdminIn = `
  SetEnvIf Authorization '^Basic +YWRtaW46' let_admin_in
  Allow from env=let_admin_in`
	}

	return fmt.Sprintf(s, authType, authName, letAdminIn, extraConfig)
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

func httpdAuthRemoteUserConf(delimiter string) string {
	s := `
RequestHeader unset X-REMOTE-USER
RequestHeader unset X-REMOTE_USER
RequestHeader unset X_REMOTE-USER
RequestHeader unset X_REMOTE_USER

RequestHeader set X_REMOTE_USER                 %%{REMOTE_USER}e           env=REMOTE_USER
RequestHeader set X_EXTERNAL_AUTH_ERROR         %%{EXTERNAL_AUTH_ERROR}e   env=EXTERNAL_AUTH_ERROR
RequestHeader set X_REMOTE_USER_EMAIL           %%{REMOTE_USER_EMAIL}e     env=REMOTE_USER_EMAIL
RequestHeader set X_REMOTE_USER_FIRSTNAME       %%{REMOTE_USER_FIRSTNAME}e env=REMOTE_USER_FIRSTNAME
RequestHeader set X_REMOTE_USER_LASTNAME        %%{REMOTE_USER_LASTNAME}e  env=REMOTE_USER_LASTNAME
RequestHeader set X_REMOTE_USER_FULLNAME        %%{REMOTE_USER_FULLNAME}e  env=REMOTE_USER_FULLNAME
RequestHeader set X_REMOTE_USER_GROUPS          %%{REMOTE_USER_GROUPS}e    env=REMOTE_USER_GROUPS
RequestHeader set X_REMOTE_USER_GROUP_DELIMITER "%s"
RequestHeader set X_REMOTE_USER_DOMAIN          %%{REMOTE_USER_DOMAIN}e    env=REMOTE_USER_DOMAIN
`
	return fmt.Sprintf(s, delimiter)
}

func uiHttpdConfig(protocol string) string {
	s := `
## ManageIQ HTTP Virtual Host Context

Listen 3000
Listen 4000

# Timeout: The number of seconds before receives and sends time out.
Timeout 120
ServerSignature Off
ServerTokens Prod

RewriteEngine On
Options SymLinksIfOwnerMatch

# LimitRequestFieldSize: Expand this to a large number to allow pass-through.
#   This does not introduce a potential DoS, because the value is validated by
#   the httpd container first.  However, if a user changes this value in the
#   httpd container, we need to be able to accomodate that value here also.
LimitRequestFieldSize 524288

# For health probes
<VirtualHost *:4000>
  RewriteRule ^/ping http://localhost:3001%%{REQUEST_URI} [P,QSA,L]
  ProxyPassReverse / http://localhost:3001/
</VirtualHost>

<VirtualHost *:3000>
  IncludeOptional conf.d/*_config

  ServerName %s://ui
  DocumentRoot /var/www/miq/vmdb/public

  RewriteCond %%{REQUEST_URI}     ^/ws/notifications [NC]
  RewriteCond %%{HTTP:UPGRADE}    ^websocket$ [NC]
  RewriteCond %%{HTTP:CONNECTION} ^Upgrade$   [NC]
  RewriteRule .* ws://localhost:3001%%{REQUEST_URI}  [P,QSA,L]
  ProxyPassReverse /ws/notifications ws://localhost:3001/ws/notifications

  RewriteRule ^/ui/service(?!/(assets|images|img|styles|js|fonts|vendor|gettext)) /ui/service/index.html [L]
  RewriteCond %%{REQUEST_URI} !^/proxy_pages
  RewriteCond %%{DOCUMENT_ROOT}/%%{REQUEST_FILENAME} !-f
  RewriteRule ^/ http://localhost:3001%%{REQUEST_URI} [P,QSA,L]
  ProxyPassReverse / http://localhost:3001/

  ProxyPreserveHost on
  <Location /assets/>
    Header unset ETag
    Header set Content-Security-Policy            "default-src 'self'; child-src 'self'; connect-src 'self'; font-src 'self' fonts.gstatic.com; script-src 'self'; style-src 'self'; report-uri /dashboard/csp_report"
    Header set X-Content-Type-Options             "nosniff"
    Header set X-Frame-Options                    "SAMEORIGIN"
    Header set X-Permitted-Cross-Domain-Policies  "none"
    Header set X-XSS-Protection                   "1; mode=block"
    FileETag None
    ExpiresActive On
    ExpiresDefault "access plus 1 year"
    Header merge Cache-Control public
  </Location>
  <Location /packs/>
    Header unset ETag
    Header set Content-Security-Policy            "default-src 'self'; child-src 'self'; connect-src 'self'; font-src 'self' fonts.gstatic.com; script-src 'self'; style-src 'self'; report-uri /dashboard/csp_report"
    Header set X-Content-Type-Options             "nosniff"
    Header set X-Frame-Options                    "SAMEORIGIN"
    Header set X-Permitted-Cross-Domain-Policies  "none"
    Header set X-XSS-Protection                   "1; mode=block"
    FileETag None
    ExpiresActive On
    ExpiresDefault "access plus 1 year"
    Header merge Cache-Control public
  </Location>
  <Location /proxy_pages/>
    ErrorDocument 403 /error/noindex.html
    ErrorDocument 404 /error/noindex.html
  </Location>
</VirtualHost>
`
	return fmt.Sprintf(s, protocol)
}

func apiHttpdConfig(protocol string) string {
	s := `
## ManageIQ HTTP Virtual Host Context

Listen 3000
Listen 4000

# Timeout: The number of seconds before receives and sends time out.
Timeout 120
ServerSignature Off
ServerTokens Prod

RewriteEngine On
Options SymLinksIfOwnerMatch

# LimitRequestFieldSize: Expand this to a large number to allow pass-through.
#   This does not introduce a potential DoS, because the value is validated by
#   the httpd container first.  However, if a user changes this value in the
#   httpd container, we need to be able to accomodate that value here also.
LimitRequestFieldSize 524288

# For health probes
<VirtualHost *:4000>
  RewriteRule ^/ping http://localhost:3001%%{REQUEST_URI} [P,QSA,L]
  ProxyPassReverse / http://localhost:3001/
</VirtualHost>

<VirtualHost *:3000>
  IncludeOptional conf.d/*_config

  ServerName %s://web-service
  DocumentRoot /var/www/miq/vmdb/public

  RewriteRule ^/ http://localhost:3001%%{REQUEST_URI} [P,QSA,L]
  ProxyPassReverse / http://localhost:3001/

  ProxyPreserveHost on
</VirtualHost>
`
	return fmt.Sprintf(s, protocol)
}

func remoteConsoleHttpdConfig(protocol string) string {
	s := `
## ManageIQ HTTP Virtual Host Context

Listen 3000
Listen 4000

# Timeout: The number of seconds before receives and sends time out.
Timeout 120
ServerSignature Off
ServerTokens Prod

RewriteEngine On
Options SymLinksIfOwnerMatch

# LimitRequestFieldSize: Expand this to a large number to allow pass-through.
#   This does not introduce a potential DoS, because the value is validated by
#   the httpd container first.  However, if a user changes this value in the
#   httpd container, we need to be able to accomodate that value here also.
LimitRequestFieldSize 524288

# For health probes
<VirtualHost *:4000>
  RewriteRule ^/ping http://localhost:3001%%{REQUEST_URI} [P,QSA,L]
  ProxyPassReverse / http://localhost:3001/
</VirtualHost>

<VirtualHost *:3000>
  IncludeOptional conf.d/*_config

  ServerName %s://remote-console
  DocumentRoot /var/www/miq/vmdb/public

  RewriteCond %%{REQUEST_URI}     ^/ws/console [NC]
  RewriteCond %%{HTTP:UPGRADE}    ^websocket$  [NC]
  RewriteCond %%{HTTP:CONNECTION} ^Upgrade$    [NC]
  RewriteRule .* ws://remote-console:3000%%{REQUEST_URI}  [P,QSA,L]
  ProxyPassReverse /ws/console ws://remote-console:3000/ws/console

  ProxyPreserveHost on
</VirtualHost>
`
	return fmt.Sprintf(s, protocol)
}

func httpdSslConfig() string {
	return `
SSLEngine on
SSLCertificateFile "/root/server.crt"
SSLCertificateKeyFile "/root/server.key"
RequestHeader set X_FORWARDED_PROTO 'https'
`
}

func appHttpdSslConfig() string {
	return `
SSLEngine on
SSLCertificateFile "/etc/pki/tls/certs/server.crt"
SSLCertificateKeyFile "/etc/pki/tls/private/server.key"
RequestHeader set X_FORWARDED_PROTO 'https'
`
}

func httpdSslProxyConfig() string {
	return `
SSLProxyEngine on
SSLProxyCACertificateFile /etc/pki/ca-trust/source/anchors/root.crt
SSLProxyCheckPeerCN on
SSLProxyCheckPeerExpire on
SSLProxyCheckPeerName on
SSLProxyVerify require
`
}
