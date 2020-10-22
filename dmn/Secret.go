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
	value string
	path  string
}

// CreateSecret create a new secret. This must be passed to the service to call the method successfully.
func (secret *Secret) CreateSecret() {

	rand.Seed(time.Now().Unix())

	var s string

	for i := 0; i < secretLength; i++ {
		random := rand.Intn(len(secretCharSet))
		randomChar := secretCharSet[random]
		s += string(randomChar)
	}
	secret.value = s
}

// SetPathToSecretsFile sets the path to the secrets file
func (secret *Secret) SetPathToSecretsFile(recmdDirPath string) {
	secret.path = filepath.Join(recmdDirPath, recmdSecretFile)
}

// WriteSecretToFile creates the file containing the secret
func (secret Secret) WriteSecretToFile() error {

	file, err := os.Create(secret.path)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.WriteString(file, secret.value)

	if err != nil {
		return err
	}

	return file.Sync()
}

// GetSecret gets the secret from the file system
func (secret Secret) GetSecret() string {
	secretData, err := ioutil.ReadFile(secret.path)

	if err != nil {
		log.Fatalf("Error, unable to read secret from file %v\n", err)
	}

	if len(secretData) != secretLength {
		log.Fatalf("Error, invalid secret length %v\n", err)
	}

	return string(secretData)
}

// Valid checks whether the secret passed in as a parameter matches our secret
func (secret Secret) Valid(test string) bool {
	return secret.value == test
}
