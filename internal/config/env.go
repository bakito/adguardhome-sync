package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/caarlos0/env/v11"

	"github.com/bakito/adguardhome-sync/internal/types"
)

// Manually collect replicas from env.
func enrichReplicasFromEnv(initialReplicas []types.Replica) ([]types.Replica, error) {
	var replicas []types.Replica
	for _, v := range os.Environ() {
		if envReplicasURLPattern.MatchString(v) {
			sm := envReplicasURLPattern.FindStringSubmatch(v)
			id, _ := strconv.Atoi(sm[1])

			if id <= 0 {
				return nil, fmt.Errorf("numbered replica env variables must have a number id >= 1, got %q", v)
			}

			if id > len(initialReplicas) {
				replicas = append(replicas, types.Replica{URL: sm[2]})
			} else {
				re := initialReplicas[id-1]
				re.URL = sm[2]
				replicas = append(replicas, re)
			}
		}
	}

	if len(replicas) == 0 {
		replicas = initialReplicas
	}

	for i := range replicas {
		reID := i + 1

		// keep the previously set value
		replicaDhcpServer := replicas[i].DHCPServerEnabled
		replicas[i].DHCPServerEnabled = nil
		prefix := fmt.Sprintf("REPLICA%d_", reID)
		if err := env.ParseWithOptions(&replicas[i], env.Options{Prefix: prefix}); err != nil {
			return nil, err
		}
		if hasReplicaFeaturesEnv(prefix) {
			if replicas[i].Features == nil {
				f := types.NewFeatures(true)
				replicas[i].Features = &f
			}
			if err := env.ParseWithOptions(replicas[i].Features, env.Options{Prefix: prefix}); err != nil {
				return nil, err
			}
		}
		if replicas[i].DHCPServerEnabled == nil {
			replicas[i].DHCPServerEnabled = replicaDhcpServer
		}
		if replicas[i].APIPath == "" {
			replicas[i].APIPath = "/control"
		}
	}

	return replicas, nil
}

func hasReplicaFeaturesEnv(prefix string) bool {
	featPrefix := prefix + "FEATURES_"
	for _, v := range os.Environ() {
		if len(v) >= len(featPrefix) && v[:len(featPrefix)] == featPrefix {
			return true
		}
	}
	return false
}
