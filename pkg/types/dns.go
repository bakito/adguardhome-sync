package types

// AccessList access list
type AccessList struct {
	AllowedClients    []string `json:"allowed_clients"`
	DisallowedClients []string `json:"disallowed_clients"`
	BlockedHosts      []string `json:"blocked_hosts"`
}

// Equals access list equal check
func (al *AccessList) Equals(o *AccessList) bool {
	return equals(al.AllowedClients, o.AllowedClients) &&
		equals(al.DisallowedClients, o.DisallowedClients) &&
		equals(al.BlockedHosts, o.BlockedHosts)
}
