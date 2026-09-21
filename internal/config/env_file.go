package config

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/bakito/adguardhome-sync/internal/types"
)

const fileSuffix = "_FILE"

var numberedReplicaPrefix = regexp.MustCompile(`^REPLICA\d+_`)

// resolveEnvFiles supports the docker secrets naming scheme. For every known config variable NAME, the variable
// NAME_FILE may point to a file whose content is used as the value of NAME. Setting both is an error.
func resolveEnvFiles() error {
	known := knownEnvNames()
	for _, kv := range os.Environ() {
		name, path, _ := strings.Cut(kv, "=")
		if !strings.HasSuffix(name, fileSuffix) || path == "" {
			continue
		}
		target := strings.TrimSuffix(name, fileSuffix)
		if !known[numberedReplicaPrefix.ReplaceAllString(target, "REPLICA_")] {
			continue
		}
		if _, ok := os.LookupEnv(target); ok {
			return fmt.Errorf("env vars %s and %s are mutually exclusive", target, name)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file of env var %s: %w", name, err)
		}
		if err := os.Setenv(target, strings.TrimRight(string(content), "\r\n")); err != nil {
			return err
		}
	}
	return nil
}

// knownEnvNames returns all env var names from the config, replicas are normalized to the REPLICA_ prefix.
func knownEnvNames() map[string]bool {
	known := make(map[string]bool)
	for _, n := range envTags(reflect.TypeFor[types.Config]()) {
		known[n] = true
	}
	for _, n := range envTags(reflect.TypeFor[types.AdGuardInstance]()) {
		known["ORIGIN_"+n] = true
		known["REPLICA_"+n] = true
	}
	for _, n := range envTags(reflect.TypeFor[types.Features]()) {
		known["REPLICA_"+n] = true
	}
	return known
}

// envTags collects the env tags of t and its nested struct fields (pointers and slices are not followed).
func envTags(t reflect.Type) []string {
	var names []string
	for f := range t.Fields() {
		if tag := f.Tag.Get("env"); tag != "" {
			names = append(names, strings.Split(tag, ",")[0])
		}
		if f.Type.Kind() == reflect.Struct {
			names = append(names, envTags(f.Type)...)
		}
	}
	return names
}
