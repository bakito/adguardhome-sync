package versions

import "golang.org/x/mod/semver"

const (
	LastStringCustomRules = "v0.107.13"
	MinAgh                = "v0.107.0"
)

func IsNewerThan(v1 string, v2 string) bool {
	return semver.Compare(v1, v2) == 1
}
