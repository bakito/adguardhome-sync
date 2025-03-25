package versions

import "golang.org/x/mod/semver"

const (
	// MinAgh minimal adguardhome version.
	MinAgh = "v0.107.40"
)

func IsNewerThan(v1, v2 string) bool {
	return semver.Compare(sanitize(v1), sanitize(v2)) == 1
}

func IsSame(v1, v2 string) bool {
	return semver.Compare(sanitize(v1), sanitize(v2)) == 0
}

func sanitize(v string) string {
	if v == "" || v[0] == 'v' {
		return v
	}
	return "v" + v
}
