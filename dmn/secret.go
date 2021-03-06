package dmn

import (
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const (
	// List of characters in our secret
	secretCharSet = "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP123456789"

	// Length of secret string
	secretLength = 40

	// The secret file
	recmdSecretFile = "recmd_secret"
)

// Secret represents the path and value of the secret. This is used to validate incoming requests.
type Secret struct {
	Value string
	Path  string
}

// Set sets the path to the secrets file
func (secret *Secret) Set(path string) {
	secret.Path = filepath.Join(path, recmdSecretFile)
}

// WriteSecretToFile creates the file containing the secret
func (secret *Secret) WriteSecretToFile() error {
	rand.Seed(time.Now().Unix())

	secret.Value = ""

	for i := 0; i < secretLength; i++ {
		random := rand.Intn(len(secretCharSet))
		randomChar := secretCharSet[random]
		secret.Value += string(randomChar)
	}

	file, err := os.Create(secret.Path)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.WriteString(file, secret.Value)

	if err != nil {
		return err
	}

	return file.Sync()
}

// GetSecret gets the secret from the file system
func (secret *Secret) GetSecret() string {
	secretData, err := ioutil.ReadFile(secret.Path)

	if err != nil {
		//log.Fatalf("In secret.GetSecret(): unable to read secret from file %v\n", err)
      log.Println("Oops, can't read secret")
	}

	if len(secretData) != secretLength {
		log.Fatalf("Error, invalid secret length %v\n", err)
	}

	return string(secretData)
}

// Valid checks whether the secret passed in as a parameter matches our secret
func (secret Secret) Valid(test string) bool {
	return secret.Value == test
}
