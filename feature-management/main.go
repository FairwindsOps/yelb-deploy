package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"time"
)

type FeatureConfig struct {
	Identifer         string    `json:"identifier"`
	AppserverImageTag string    `json:"appserverImageTag"`
	UiImageTag        string    `json:"uiImageTag"`
	LastDeployed      time.Time `json:"lastDeployed"`
}

const usageString = `
feature-management

A script to manage feature branch deployments in this repository

Requires a 'prune' or 'generate' subcommand
`

func main() {
	pruneCmd := flag.NewFlagSet("prune", flag.ExitOnError)
	pruneCmdMaxFeatures := pruneCmd.Int("max-features", 2, "the maximum number of feature branches to keep")

	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	generateCmdBranchName := generateCmd.String("branch", "", "the branch name being deployed")
	generateCmdAppserverTag := generateCmd.String("appserverTag", "main", "the tag to use for the appserver")
	generateCmdUITag := generateCmd.String("ui tag", "main", "the tag to use for the ui")

	if len(os.Args) < 2 {
		fmt.Printf(usageString)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "prune":
		pruneCmd.Parse(os.Args[2:])
		log.Printf("prune command:\nmaxFeatures: %d", *pruneCmdMaxFeatures)
		allFeatures, err := getAllFeatures()
		if err != nil {
			log.Fatal(err)
		}

		if err := prune(allFeatures, *pruneCmdMaxFeatures); err != nil {
			log.Fatal(err)
		}
	case "generate":
		generateCmd.Parse(os.Args[2:])

		if *generateCmdBranchName == "" {
			branch := os.Getenv("CIRCLE_BRANCH")
			if branch == "" {
				log.Fatal("could not determine the branch name")
			}
			generateCmdBranchName = &branch
		}

		if err := generateFeatureConfig(*generateCmdBranchName, *generateCmdUITag, *generateCmdAppserverTag); err != nil {
			log.Fatal(err)
		}

	default:
		fmt.Printf(usageString)
		os.Exit(1)
	}
}

func generateFeatureConfig(branchName string, uiTag string, appserverTag string) error {
	sanitizedBranchName := sanitizeBranchName(branchName)

	log.Printf(`
generate command args:
identifier: %s
appserverTag: %s
uiTag: %s
`, sanitizedBranchName, appserverTag, uiTag)

	newFeatureConfig := FeatureConfig{
		AppserverImageTag: appserverTag,
		UiImageTag:        uiTag,
		LastDeployed:      time.Now(),
		Identifer:         sanitizedBranchName,
	}

	outputJSON, err := json.MarshalIndent(newFeatureConfig, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("writing new feature config: %v\n", string(outputJSON))
	fileName := fmt.Sprintf("feature/%s.json", sanitizedBranchName)

	f, _ := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	f.WriteString(string(outputJSON))

	return nil
}

func sanitizeBranchName(in string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9 ]+")
	return reg.ReplaceAllString(in, "-")
}

// getAllFeatures will read the feature/ folder and
// return a list of all the FeatureConfig structs
// that are in that folder
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
		return allFeatures[i].LastDeployed.After(allFeatures[j].LastDeployed)
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
