package main

import (
	"os"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestConfigCreate(t *testing.T) {
	config_file := "config.yaml"
	file, err := os.Open(config_file)
	if err != nil {
		t.Errorf("Error opening file: %s", err)
	}

	var config ConfigBasic
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		t.Errorf("Error decoding file: %s", err)
	}

	if len(config.Services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(config.Services))
	}

	if len(config.Networks) != 1 {
		t.Errorf("Expected 1 services, got %d", len(config.Networks))
	}
}
