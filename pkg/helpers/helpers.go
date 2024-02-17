package helpers

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

func SanitizePath(path string) string {
	var sPath string

	usr, _ := user.Current()
	dir := usr.HomeDir

	if path == "~" || path == "$HOME" {
		sPath = dir
	} else if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "$HOME/") {
		sPath = filepath.Join(dir, path[2:])
	} else {
		sPath = path
	}

	return sPath
}

func FileExists(path string) bool {
	if info, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else {
		return !info.IsDir()
	}
}

func PathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func GetFileNameWithoutExtension(path string) string {
	return strings.Split(GetFileName(path), ".")[0]
}

func GetFileName(path string) string {
	return strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
}

func GetFilesInDir(path string) ([]string, error) {
	var files []string

	if !PathExists(path) {
		return files, fmt.Errorf("path '%s' does not exist", path)
	}

	// recursivly search for images in provided path
	err := filepath.WalkDir(path, func(path string, dir fs.DirEntry, err error) error {
		if !dir.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return files, err
	}

	return files, nil
}

func CreateHash(str string) string {
	hash := sha256.New()
	hash.Write([]byte(str))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func WriteToFile(path string, data string) error {
	return os.WriteFile(path, []byte(data), 0644)
}

func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot read data from file %s, %+v", path, err)
	}

	return data, nil
}

func ValidateInputLength(input string, minLen, maxLen int) bool {
	if len(input) < minLen || len(input) > maxLen {
		return false
	}
	return true
}

func IsValidIpv4(host string) bool {
	// NOTE: thanks to: https://gist.github.com/JayGoldberg/4a2a0d159342e53434f2785d7c9cbec5
	parts := strings.Split(host, ".")

	if len(parts) < 4 {
		return false
	}

	for _, x := range parts {
		if i, err := strconv.Atoi(x); err == nil {
			if i < 0 || i > 255 {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func IsValidUrl(uri string) bool {
	if _, err := url.ParseRequestURI(uri); err != nil {
		return false
	}
	return true
}
