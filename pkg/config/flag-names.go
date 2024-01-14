package config

const (
	FlagCron            = "cron"
	FlagRunOnStart      = "runOnStart"
	FlagPrintConfigOnly = "printConfigOnly"
	FlagContinueOnError = "continueOnError"

	FlagApiPort     = "api-port"
	FlagApiUsername = "api-username"
	FlagApiPassword = "api-password"
	FlagApiDarkMode = "api-dark-mode"

	FlagFeatureDhcpServerConfig = "feature-dhcp-server-config"
	FlagFeatureDhcpStaticLeases = "feature-dhcp-static-leases"
	FlagFeatureDnsServerConfig  = "feature-dns-server-config"
	FlagFeatureDnsAccessLists   = "feature-dns-access-lists"
	FlagFeatureDnsRewrites      = "feature-dns-rewrites"
	FlagFeatureGeneral          = "feature-general-settings"
	FlagFeatureQueryLog         = "feature-query-log-config"
	FlagFeatureStats            = "feature-stats-config"
	FlagFeatureClient           = "feature-client-settings"
	FlagFeatureServices         = "feature-services"
	FlagFeatureFilters          = "feature-filters"

	FlagOriginURL      = "origin-url"
	FlagOriginWebURL   = "origin-web-url"
	FlagOriginApiPath  = "origin-api-path"
	FlagOriginUsername = "origin-username"

	FlagOriginPassword = "origin-password"
	FlagOriginCookie   = "origin-cookie"
	FlagOriginISV      = "origin-insecure-skip-verify"

	FlagReplicaURL           = "replica-url"
	FlagReplicaWebURL        = "replica-web-url"
	FlagReplicaApiPath       = "replica-api-path"
	FlagReplicaUsername      = "replica-username"
	FlagReplicaPassword      = "replica-password"
	FlagReplicaCookie        = "replica-cookie"
	FlagReplicaISV           = "replica-insecure-skip-verify"
	FlagReplicaAutoSetup     = "replica-auto-setup"
	FlagReplicaInterfaceName = "replica-interface-name"
)
