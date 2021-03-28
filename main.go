package main

import (
	"github.com/bakito/adguardhome-sync/pkg/sync"
	"os"

	"github.com/bakito/adguardhome-sync/pkg/log"

	"github.com/bakito/adguardhome-sync/pkg/client"
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

	origin, err := client.New(os.Getenv(envOriginApiURL), os.Getenv(envOriginUsername), os.Getenv(envOriginPassword))
	if err != nil {
		panic(err)
	}
	replica, err := client.New(os.Getenv(envReplicaApiURL), os.Getenv(envReplicaUsername), os.Getenv(envOReplicaPassword))
	if err != nil {
		panic(err)
	}
	if err = sync.Sync(origin, replica); err != nil {
		panic(err)
	}
}
