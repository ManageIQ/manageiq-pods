package miqtools

func postgresqlOverrideConf() string {
	return `
#------------------------------------------------------------------------------
# CONNECTIONS AND AUTHENTICATION
#------------------------------------------------------------------------------

tcp_keepalives_count = 9
tcp_keepalives_idle = 3
tcp_keepalives_interval = 75

password_encryption = scram-sha-256

#------------------------------------------------------------------------------
# RESOURCE USAGE (except WAL)
#------------------------------------------------------------------------------

max_worker_processes = 10

#------------------------------------------------------------------------------
# WRITE AHEAD LOG
#------------------------------------------------------------------------------

wal_level = 'logical'
wal_log_hints = on
wal_buffers = 16MB
checkpoint_completion_target = 0.9

#------------------------------------------------------------------------------
# REPLICATION
#------------------------------------------------------------------------------

max_wal_senders = 10
wal_sender_timeout = 0
max_replication_slots = 10
hot_standby = on

#------------------------------------------------------------------------------
# ERROR REPORTING AND LOGGING
#------------------------------------------------------------------------------

log_destination = 'stderr'
# This is used when logging to stderr:
logging_collector = off

log_min_duration_statement = 5000
log_connections = on
log_disconnections = on
log_line_prefix = '%t:%r:%c:%u@%d:[%p]:'
log_lock_waits = on

#------------------------------------------------------------------------------
# AUTOVACUUM PARAMETERS
#------------------------------------------------------------------------------

log_autovacuum_min_duration = 0
autovacuum_naptime = 5min
autovacuum_vacuum_threshold = 500
autovacuum_analyze_threshold = 500
autovacuum_vacuum_scale_factor = 0.05

#------------------------------------------------------------------------------
# RESOURCE USAGE (except WAL)
#------------------------------------------------------------------------------

dynamic_shared_memory_type = sysv
                                        # the default is the first option
                                        # supported by the operating system:
                                        #   posix
                                        #   sysv
                                        #   windows
                                        #   mmap
                                        # (change requires restart)

#------------------------------------------------------------------------------
# LOCK MANAGEMENT
#------------------------------------------------------------------------------

deadlock_timeout = 5s

#------------------------------------------------------------------------------
# VERSION/PLATFORM COMPATIBILITY
#------------------------------------------------------------------------------

escape_string_warning = off
standard_conforming_strings = off
`
}

func postgresqlSslConf() string {
	return `
#------------------------------------------------------------------------------
# SSL CONFIG
#------------------------------------------------------------------------------

ssl = on
ssl_cert_file = '/etc/pki/tls/certs/server.crt' # server certificate
ssl_key_file =  '/etc/pki/tls/private/server.key' # server private key
#ssl_ca_file                                   # trusted certificate authorities
#ssl_crl_file                                  # certificates revoked by certificate authorities

`
}
