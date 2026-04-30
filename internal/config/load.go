// Package config loads and validates job configuration files.
package config

import (
	"context"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(_ context.Context, path string) (*Job, error) {
	if path == "" {
		return nil, errors.New("config path is required (-c <job.yaml>)")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}

		return nil, fmt.Errorf("check config file: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("config path points to directory: %s", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config file: %w", err)
	}
	defer file.Close()

	var job Job
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&job); err != nil {
		return nil, fmt.Errorf("decode job config: %w", err)
	}

	if err := Validate(job); err != nil {
		return nil, err
	}

	return &job, nil
}
