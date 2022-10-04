package versions

import "golang.org/x/mod/semver"

const (
	// MinAgh minimal adguardhome version
	MinAgh = "v0.107.0"
	// LastStringCustomRules last adguardhome version with string payload custom rules
	// https://github.com/bakito/adguardhome-sync/issues/99
	LastStringCustomRules = "v0.107.13"
	// IncompatibleAPI adguardhome  version with incompatible API
	// https://github.com/bakito/adguardhome-sync/issues/99
	IncompatibleAPI = "v0.107.14"
	// FixedIncompatibleAPI adguardhome version with fixed API
	// https://github.com/bakito/adguardhome-sync/issues/99
	FixedIncompatibleAPI = "v0.107.15"
)

func IsNewerThan(v1 string, v2 string) bool {
	return semver.Compare(v1, v2) == 1
}

func IsSame(v1 string, v2 string) bool {
	return semver.Compare(v1, v2) == 0
}
