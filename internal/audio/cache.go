package audio

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
)

func GenerateHash(text, lang, voice, format string) string {
	hashInput := fmt.Sprintf("%s|%s|%s|%s", text, lang, voice, format)
	hashBytes := md5.Sum([]byte(hashInput))
	return fmt.Sprintf("%x", hashBytes)
}

func GetPaths(cacheDir, hash, format string) (pathNoExt, fullPath string) {
	pathNoExt = filepath.Join(cacheDir, hash)
	fullPath = pathNoExt + "." + format
	return pathNoExt, fullPath
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
