package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/bakito/adguardhome-sync/internal/versions"
)

func main() {
	targetVersion := ""
	if len(os.Args) > 1 {
		targetVersion = os.Args[1]
	}
	if targetVersion == "" {
		log.Fatal("AdGuardHome target version argument required")
	}

	valuesFile := findValuesFile()
	if err := updateReplicaVersions(valuesFile, versions.MinAgh, targetVersion, "latest"); err != nil {
		log.Fatalf("Failed to update replica versions in %s: %v", valuesFile, err)
	}
	log.Printf("Successfully updated replica versions in %s (%s, %s, latest)", valuesFile, versions.MinAgh, targetVersion)
}

func findValuesFile() string {
	candidates := []string{
		"testdata/e2e/values.yaml",
		"../../testdata/e2e/values.yaml",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return "testdata/e2e/values.yaml"
}

func updateReplicaVersions(filePath, minVersion, targetVersion, latestVersion string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("unmarshal yaml: %w", err)
	}

	// Update replica.versions in the YAML AST to preserve structure
	if !modifyReplicaVersionsNode(&root, minVersion, targetVersion, latestVersion) {
		return fmt.Errorf("replica.versions node not found in %s", filePath)
	}

	outFile, err := os.OpenFile(filepath.Clean(filePath), os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("open file for writing: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	encoder := yaml.NewEncoder(outFile)
	encoder.SetIndent(2)
	if err := encoder.Encode(&root); err != nil {
		return fmt.Errorf("encode yaml: %w", err)
	}

	return nil
}

func modifyReplicaVersionsNode(root *yaml.Node, minVersion, targetVersion, latestVersion string) bool {
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return false
	}
	doc := root.Content[0]
	for i := 0; i < len(doc.Content); i += 2 {
		if doc.Content[i].Value != "replica" {
			continue
		}
		replicaMap := doc.Content[i+1]
		for j := 0; j < len(replicaMap.Content); j += 2 {
			if replicaMap.Content[j].Value != "versions" {
				continue
			}
			versionsSeq := replicaMap.Content[j+1]
			versionsSeq.Kind = yaml.SequenceNode
			versionsSeq.Content = []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: minVersion},
				{Kind: yaml.ScalarNode, Value: targetVersion},
				{Kind: yaml.ScalarNode, Value: latestVersion},
			}
			return true
		}
	}
	return false
}
