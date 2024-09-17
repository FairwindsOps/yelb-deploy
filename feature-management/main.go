package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
)

type FeatureConfig struct {
	Identifer         string `json:"identifier"`
	AppserverImageTag string `json:"appserverImageTag"`
	UiImageTag        string `json:"uiImageTag"`
	LastDeployed      string `json:"lastDeployed"`
}

const usageString = `
Expected 'prune' or 'generate' subcommands
`

func main() {
	pruneCmd := flag.NewFlagSet("prune", flag.ExitOnError)
	pruneCmdMaxFeatures := pruneCmd.Int("max-features", 6, "the maximum number of feature branches to keep")

	if len(os.Args) < 2 {
		fmt.Println(usageString)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "prune":
		pruneCmd.Parse(os.Args[2:])
		fmt.Println("subcommand 'parse'")
		fmt.Println("  maxFeatures:", *pruneCmdMaxFeatures)
		allFeatures, err := getAllFeatures()
		if err != nil {
			log.Fatal(err)
		}

		if err := prune(allFeatures, *pruneCmdMaxFeatures); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Println(usageString)
		os.Exit(1)
	}

}

func getAllFeatures() ([]FeatureConfig, error) {
	allFeatures := []FeatureConfig{}
	var allErrors error

	items, _ := os.ReadDir("./feature")
	for _, item := range items {
		if item.IsDir() {
			log.Printf("skipping directory %s", item.Name())
			continue
		}
		if item.Name() == ".gitkeep" {
			continue
		}

		content, err := os.ReadFile(fmt.Sprintf("feature/%s", item.Name()))
		if err != nil {
			allErrors = errors.Join(allErrors, err)
			continue
		}

		feature := FeatureConfig{}
		if err := json.Unmarshal(content, &feature); err != nil {
			allErrors = errors.Join(err)
			continue
		}
		output, err := json.MarshalIndent(feature, "", "  ")
		if err != nil {
			allErrors = errors.Join(allErrors, err)
			continue
		}
		log.Printf("%s\n%s", item.Name(), string(output))

		allFeatures = append(allFeatures, feature)
	}
	return allFeatures, allErrors
}

// prune will delete the json file for any features
// that are above the maximum count, starting with
// the oldest lastDeployed
func prune(allFeatures []FeatureConfig, maxFeatures int) error {
	if len(allFeatures) < maxFeatures {
		log.Println("total features is less than maxFeatures - nothing to do")
		return nil
	}

	var totalErrors error
	sort.Slice(allFeatures, func(i, j int) bool {
		return allFeatures[i].LastDeployed < allFeatures[j].LastDeployed
	})

	featuresToKeep := allFeatures[:maxFeatures]
	featuresToRemove := allFeatures[maxFeatures:]

	log.Printf("keep: %v", featuresToKeep)
	log.Printf("remove: %v", featuresToRemove)

	for _, feature := range featuresToRemove {
		err := os.Remove("feature/" + feature.Identifer + ".json")
		if err != nil {
			err = errors.Join(totalErrors, err)
		}
	}
	return totalErrors
}
