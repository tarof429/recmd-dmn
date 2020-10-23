package dmn

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	TestataDirPath = "testdata"
)

var (
	TestPath    string
	TestDataDir string
	TestHistory HistoryFile
	TestSecret  Secret
)

// InitHandlerTest deletes and recreates the testdata sandbox directory for testing.
// Global variables used by the test are also initialized.
func InitHandlerTest() error {

	fmt.Println("Intializing testdata dir...")

	TestPath, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	TestDataDir = filepath.Join(TestPath, TestataDirPath)
	os.RemoveAll(TestDataDir)

	mode := os.FileMode(0755)
	os.Mkdir(TestDataDir, mode)

	// Write the secret
	TestSecret.Set(TestDataDir)
	err = TestSecret.WriteSecretToFile()
	if err != nil {
		return err
	}
	if TestSecret.GetSecret() == "" {
		return err
	}

	// Create a history file
	TestHistory.Set(TestDataDir)
	err = TestHistory.WriteHistoryToFile()
	if err != nil {
		return err
	}
	return nil

}
