package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/caarlos0/env/v11"

	"github.com/bakito/adguardhome-sync/internal/types"
)

// Manually collect replicas from env.
func enrichReplicasFromEnv(initialReplicas []types.AdGuardInstance) ([]types.AdGuardInstance, error) {
	var replicas []types.AdGuardInstance
	for _, v := range os.Environ() {
		if envReplicasURLPattern.MatchString(v) {
			sm := envReplicasURLPattern.FindStringSubmatch(v)
			id, _ := strconv.Atoi(sm[1])

			if id <= 0 {
				return nil, fmt.Errorf("numbered replica env variables must have a number id >= 1, got %q", v)
			}

			if id > len(initialReplicas) {
				replicas = append(replicas, types.AdGuardInstance{URL: sm[2]})
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
		if err := env.ParseWithOptions(&replicas[i], env.Options{Prefix: fmt.Sprintf("REPLICA%d_", reID)}); err != nil {
			return nil, err
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
