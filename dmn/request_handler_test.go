package dmn

import (
	"fmt"
)

const (
	TestataDirPath = "testdata"
)

var (
	TestApp App
	// TestPath    string
	// TestDataDir string
	// TestLogFile LogFile
	// TestHistory HistoryFile
	// TestSecret  Secret
)

// InitHandlerTest deletes and recreates the testdata sandbox directory for testing.
// Global variables used by the test are also initialized.
func (TestApp *App) InitHandlerTest() error {

	fmt.Println("Intializing testdata dir...")

	footprint := Footprint{}
	footprint.TestFootprint()

	TestApp.Footprint.Set(&footprint)

	// Set the secret file
	TestApp.DmnSecret.Set(footprint.confDirPath)
	err := TestApp.DmnSecret.WriteSecretToFile()
	if err != nil {
		return err
	}
	if TestApp.DmnSecret.GetSecret() == "" {
		return err
	}

	// Set the log file
	TestApp.DmnLogFile.Set(footprint.logDirPath)
	TestApp.DmnLogFile.Create()

	// Set the history file
	TestApp.History.Set(footprint.confDirPath)
	TestApp.History.Remove()
	err = TestApp.History.WriteHistoryToFile()
	if err != nil {
		return err
	}

	TestApp.RequestHandler.Set(TestApp.DmnSecret,
		TestApp.History, TestApp.DmnLogFile.Log)

	return nil

}
