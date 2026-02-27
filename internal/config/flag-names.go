package config

const (
	FlagCron            = "cron"
	FlagRunOnStart      = "runOnStart"
	FlagPrintConfigOnly = "printConfigOnly"
	FlagContinueOnError = "continueOnError"

	FlagAPIPort     = "api-port"
	FlagAPIUsername = "api-username"
	FlagAPIPassword = "api-password"
	FlagAPIDarkMode = "api-dark-mode"

	FlagFeatureDhcpServerConfig = "feature-dhcp-server-config"
	FlagFeatureDhcpStaticLeases = "feature-dhcp-static-leases"
	FlagFeatureDNSServerConfig  = "feature-dns-server-config"
	FlagFeatureDNSAccessLists   = "feature-dns-access-lists"
	FlagFeatureDNSRewrites      = "feature-dns-rewrites"
	FlagFeatureGeneral          = "feature-general-settings"
	FlagFeatureQueryLog         = "feature-query-log-config"
	FlagFeatureStats            = "feature-stats-config"
	FlagFeatureClient           = "feature-client-settings"
	FlagFeatureServices         = "feature-services"
	FlagFeatureFilters          = "feature-filters"
	FlagFeatureTLSConfig        = "feature-tls-config"
	FlagFeatureProtectionStatus = "feature-protection_status"

	FlagOriginURL      = "origin-url"
	FlagOriginWebURL   = "origin-web-url"
	FlagOriginAPIPath  = "origin-api-path"
	FlagOriginUsername = "origin-username"

	FlagOriginPassword = "origin-password"
	FlagOriginCookie   = "origin-cookie"
	FlagOriginISV      = "origin-insecure-skip-verify"

	FlagReplicaURL           = "replica-url"
	FlagReplicaWebURL        = "replica-web-url"
	FlagReplicaAPIPath       = "replica-api-path"
	FlagReplicaUsername      = "replica-username"
	FlagReplicaPassword      = "replica-password"
	FlagReplicaCookie        = "replica-cookie"
	FlagReplicaISV           = "replica-insecure-skip-verify"
	FlagReplicaAutoSetup     = "replica-auto-setup"
	FlagReplicaInterfaceName = "replica-interface-name"
)
