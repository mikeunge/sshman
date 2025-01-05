package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

func SanitizePath(path string) string {
	sPath := path

	usr, _ := user.Current()
	dir := usr.HomeDir

	if path == "~" || path == "$HOME" {
		sPath = dir
	} else if strings.HasPrefix(path, "~/") {
		sPath = filepath.Join(dir, path[2:])
	} else if strings.HasPrefix(path, "$HOME/") {
		sPath = filepath.Join(dir, path[5:])
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

func PathIsFile(path string) (bool, error) {
	var err error
	var info os.FileInfo

	if info, err = os.Stat(path); err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, nil
	}
	return true, nil
}

func CreatePathIfNotExist(path string) error {
	// check if we deal with a path or a filepath
	if len(strings.Split(GetFileName(path), ".")) > 1 {
		path = strings.Join(strings.Split(path, "/")[:len(strings.Split(path, "/"))-1], "/")
	}
	if PathExists(path) {
		return nil
	}
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func RemovePath(path string) error {
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
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

	// recursivly search for files in provided path
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

func WriteToFile(path string, data string, perm os.FileMode) error {
	return os.WriteFile(path, []byte(data), perm)
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

func IsValidIp(host string) bool {
	if ip := net.ParseIP(host); ip == nil {
		return false
	}
	return true
}

func IsValidUrl(uri string) bool {
	if uri == "localhost" {
		return true
	}

	re := regexp.MustCompile(`^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/|\/|\/\/)?[A-z0-9_-]*?[:]?[A-z0-9_-]*?[@]?[A-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)
	return re.MatchString(uri)
}

func EncryptString(data string, encKey string) (string, error) {
	//Since the key is in string, we need to convert decode it to bytes
	key, _ := hex.DecodeString(encKey)
	plaintext := []byte(data)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	//Encrypt the data using aesGCM.Seal

	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

func DecryptString(data string, encKey string) (string, error) {
	key, _ := hex.DecodeString(encKey)
	enc, _ := hex.DecodeString(data)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
