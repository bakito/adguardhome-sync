package main

import (
	"github.com/bakito/adguardhome-sync/pkg/client"
	"os"
)

const (
	envOriginApiURL     = "ORIGIN_API_URL"
	envOriginUsername   = "ORIGIN_USERNAME"
	envOriginPassword   = "ORIGIN_PASSWORD"
	envReplicaApiURL    = "REPLICA_API_URL"
	envReplicaUsername  = "REPLICA_USERNAME"
	envOReplicaPassword = "REPLICA_PASSWORD"
)

func main() {
	// Create a Resty Client

	origin, err := client.New(os.Getenv(envOriginApiURL), os.Getenv(envOriginUsername), os.Getenv(envOriginPassword))
	if err != nil {
		panic(err)
	}
	replica, err := client.New(os.Getenv(envReplicaApiURL), os.Getenv(envReplicaUsername), os.Getenv(envOReplicaPassword))
	if err != nil {
		panic(err)
	}

	err = syncRewrites(err, origin, replica)
	if err != nil {
		panic(err)
	}
	err = syncFilters(err, origin, replica)
	if err != nil {
		panic(err)
	}

	// POST http://192.168.2.207/control/dns_config {"protection_enabled":false}
}

func syncFilters(err error, origin client.Client, replica client.Client) error {
	of, err := origin.Filtering()
	if err != nil {
		return err
	}
	rf, err := replica.Filtering()
	if err != nil {
		return err
	}

	fa, fd := rf.Filters.Merge(of.Filters)

	if err = replica.AddFilters(false, fa...); err != nil {
		return err
	}

	if len(fa) > 0 {
		if err = replica.RefreshFilters(false); err != nil {
			return err
		}
	}

	if err = replica.DeleteFilters(false, fd...); err != nil {
		return err
	}

	fa, fd = rf.WhitelistFilters.Merge(of.WhitelistFilters)
	if err = replica.AddFilters(true, fa...); err != nil {
		return err
	}

	if len(fa) > 0 {
		if err = replica.RefreshFilters(true); err != nil {
			return err
		}
	}

	if err = replica.DeleteFilters(true, fd...); err != nil {
		return err
	}

	if of.UserRules.String() != rf.UserRules.String() {
		return replica.SetCustomRules(of.UserRules)
	}

	if of.Enabled != rf.Enabled || of.Interval != rf.Interval {
		if err = replica.ToggleFiltering(of.Enabled, of.Interval); err != nil {
			return err
		}
	}
	return nil
}

func syncRewrites(err error, origin client.Client, replica client.Client) error {
	originRewrites, err := origin.RewriteList()
	if err != nil {
		return err
	}
	replicaRewrites, err := replica.RewriteList()
	if err != nil {
		return err
	}

	a, r := replicaRewrites.Merge(originRewrites)

	if err = replica.AddRewriteEntries(a...); err != nil {
		return err
	}
	if err = replica.DeleteRewriteEntries(r...); err != nil {
		return err
	}
	return nil
}
