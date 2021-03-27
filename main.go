package main

import (
	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/log"
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

var (
	l = log.GetLogger("main")
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

	// POST http://192.168.2.207/control/filtering/config {"interval":24,"enabled":false}
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

	err = replica.AddFilters(false, fa...)
	if err != nil {
		return err
	}

	if len(fa) > 0 {
		err = replica.RefreshFilters(false)
		if err != nil {
			return err
		}
	}

	err = replica.DeleteFilters(false, fd...)
	if err != nil {
		return err
	}

	fa, fd = rf.WhitelistFilters.Merge(of.WhitelistFilters)
	err = replica.AddFilters(true, fa...)
	if err != nil {
		return err
	}

	if len(fa) > 0 {
		err = replica.RefreshFilters(true)
		if err != nil {
			return err
		}
	}

	err = replica.DeleteFilters(true, fd...)
	if err != nil {
		return err
	}

	if of.UserRules.String() != rf.UserRules.String() {
		return replica.SetCustomRules(of.UserRules)
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

	err = replica.AddRewriteEntries(a...)
	if err != nil {
		return err
	}
	err = replica.DeleteRewriteEntries(r...)
	if err != nil {
		return err
	}
	return err
}
