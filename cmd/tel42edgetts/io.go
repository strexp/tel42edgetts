package main

import (
	"os"
)

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func removeFile(path string) error {
	return os.Remove(path)
}
