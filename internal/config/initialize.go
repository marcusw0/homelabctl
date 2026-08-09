package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func Initialize(configPath string) error {

	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	file, err := os.OpenFile(
		configPath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0o600,
	)

	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("config file already exists: %s", configPath)
	}

	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close config file: %w", err)
	}
	return nil
}
