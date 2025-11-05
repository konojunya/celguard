package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func resolveConfigPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if gw := os.Getenv("GITHUB_WORKSPACE"); gw != "" {
		return filepath.Join(gw, path)
	}
	return path
}

func ReadConfig() (*Config, error) {
	path := ".github/prlint.yaml"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	b, err := os.ReadFile(resolveConfigPath(path))
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadEventFromGitHub() (*Event, error) {
	evPath := os.Getenv("GITHUB_EVENT_PATH")
	if evPath == "" {
		return nil, fmt.Errorf("GITHUB_EVENT_PATH is not set")
	}

	ev, err := os.ReadFile(evPath)
	if err != nil {
		return nil, err
	}

	var event Event
	err = json.Unmarshal(ev, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}
