package fileparser

import (
	"encoding/json"
	"envoy-cp/processors/models"
	"fmt"
	"io/ioutil"
	"log"
)



func ParseJson(filePath string) (*models.EnvoyConfig, error) {
	log.Println("Parsing file " + filePath)

	var envoyConfig models.EnvoyConfig

	jsonFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error reading YAML file: %s", err)
	}

	err = json.Unmarshal(jsonFile, &envoyConfig)
	if err != nil {
		return nil, err
	}

	return &envoyConfig, nil
}