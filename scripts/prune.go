package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
)

var maxFeatures int = 6

type FeatureConfig struct {
	Identifer         string `json:"identifier"`
	AppserverImageTag string `json:"appserverImageTag"`
	UiImageTag        string `json:"uiImageTag"`
	LastDeployed      string `json:"lastDeployed"`
}

func main() {

	allFeatures := []FeatureConfig{}

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
			log.Fatalf("error when opening file: %s", err.Error())
		}

		feature := FeatureConfig{}
		if err := json.Unmarshal(content, &feature); err != nil {
			log.Fatalf("error reading feature: %s", err.Error())
		}
		output, err := json.MarshalIndent(feature, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s\n%s", item.Name(), string(output))

		allFeatures = append(allFeatures, feature)
	}

	sort.Slice(allFeatures, func(i, j int) bool {
		return allFeatures[i].LastDeployed < allFeatures[j].LastDeployed
	})

	featuresToKeep := allFeatures[:maxFeatures]
	featuresToRemove := allFeatures[maxFeatures:]

	log.Printf("keep: %v", featuresToKeep)
	log.Printf("remove: %v", featuresToRemove)

	for _, feature := range featuresToRemove {
		os.Remove("feature/" + feature.Identifer + ".json")
	}
}
