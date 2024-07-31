package utils

import (
	"log"
	"os"
	"path/filepath"
)

func PathJoinWithHome(path string) string {
	dir, err := os.UserHomeDir()

	if err != nil {
		log.Fatalf("Error getting home dir: %v", err)
	}

	return filepath.Join(dir, path)
}
