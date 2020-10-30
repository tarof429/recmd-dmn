package main

/*
Copyright © 2020 Taro Fukunaga <tarof429@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	dmn "github.com/tarof429/recmd-dmn/dmn"
)

const (
	// DefaultServerPort is the port that this program listens to
	DefaultServerPort = ":8999"

	// DefaultConfigDir is a directory containing configuration and the Command history file
	DefaultConfigDir = ".recmd"

	// TestConfigDir is used for testing
	TestConfigDir = "testdata"

	DefaultLogFile = "recmd-dmn.log"
)

// App represents this API server
type App struct {
	ConfigPath     string
	Router         *mux.Router
	RequestHandler dmn.RequestHandler
	DmnSecret      dmn.Secret
}

// Initialize initializes the application
func (a *App) Initialize(configPath string) {

	a.CreateLogs()

	a.Router = mux.NewRouter()

	a.InitializeConfigPath(configPath)

	secret := a.CreateSecret()

	historyFile := a.CreateHistoryFile()

	a.RequestHandler.Set(secret, historyFile)

	a.InitializeRoutes()

	a.RequestHandler.CommandScheduler.CreateScheduler()
	go a.RequestHandler.CommandScheduler.RunScheduler()
}

// InitializeConfigPath creates the config directory if it doesn't exist
func (a *App) InitializeConfigPath(configPath string) {

	a.ConfigPath = configPath

	fileInfo, statErr := os.Stat(a.ConfigPath)

	if os.IsNotExist(statErr) {

		mode := int(0755)
		err := os.Mkdir(a.ConfigPath, os.FileMode(mode))

		if err != nil {
			log.Fatalf("Error, unable to create ~/.recmd: %v\n", err)
		}
	} else if !fileInfo.IsDir() {
		log.Fatalf("Error, ~/.recmd is not a directory")
	}
}

// CreateLogs creates a new log file
func (a *App) CreateLogs() {
	f, err := os.Create(DefaultLogFile)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	// Not sure how to close it
	//defer f.Close()

	a.RequestHandler.Log = log.New(f, "", log.LstdFlags|log.Lshortfile)

}

// CreateSecret creates the secret whenever the application starts
func (a *App) CreateSecret() dmn.Secret {

	a.DmnSecret.Set(a.ConfigPath)

	err := a.DmnSecret.WriteSecretToFile()

	if err != nil {
		log.Printf("Error, unable to create secrets file %v\n", err)

	}

	if a.DmnSecret.GetSecret() == "" {
		log.Printf("Error, secret was an empty string")
	}

	return a.DmnSecret
}

// GetSecret just returns the secret
func (a *App) GetSecret() dmn.Secret {
	return a.DmnSecret
}

// CreateHistoryFile initializes the historyFile file
func (a *App) CreateHistoryFile() dmn.HistoryFile {

	var historyFile dmn.HistoryFile

	historyFile.Set(a.ConfigPath)

	_, statErr := os.Stat(historyFile.Path)

	if os.IsNotExist(statErr) {

		err := historyFile.WriteHistoryToFile()

		if err != nil {
			log.Printf("Error, unable to create historyFile file")

		}
	}

	return historyFile
}

// InitializeRoutes initializes the routes for this application
func (a *App) InitializeRoutes() {
	a.Router.HandleFunc("/secret/{secret}/delete/cmdHash/{cmdHash}", a.RequestHandler.HandleDelete)
	a.Router.HandleFunc("/secret/{secret}/add/command/{command}/description/{description}/workingDirectory/{workingDirectory}", a.RequestHandler.HandleAdd)
	a.Router.HandleFunc("/secret/{secret}/select/cmdHash/{cmdHash}", a.RequestHandler.HandleSelect)
	a.Router.HandleFunc("/secret/{secret}/search/description/{description}", a.RequestHandler.HandleSearch)
	a.Router.HandleFunc("/secret/{secret}/run/cmdHash/{cmdHash}", a.RequestHandler.HandleRun)
	a.Router.HandleFunc("/secret/{secret}/list", a.RequestHandler.HandleList)
	a.Router.HandleFunc("/secret/{secret}/status", a.RequestHandler.HandleStatus)

	http.Handle("/", a.Router)
}

// Run runs the application
func (a *App) Run() {
	log.Printf("Starting server on %v\n", DefaultServerPort)
	log.Fatal(http.ListenAndServe(DefaultServerPort, nil))
}

// GetDefaultConfigPath gets the default configPath which is ~/.recmd
func GetDefaultConfigPath() string {

	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatalf("Error, unable to obtain home directory path %v\n", err)
	}

	return filepath.Join(homeDir, DefaultConfigDir)
}

// GetTestConfigPath gets the test configPath which is ./testdata
func GetTestConfigPath() string {

	testPath, err := os.Getwd()

	if err != nil {
		log.Println(err)
	}

	return filepath.Join(testPath, TestConfigDir)
}

func main() {

	a := App{}

	configPath := GetDefaultConfigPath()

	a.Initialize(configPath)

	a.RequestHandler.Log.Printf("Starting up!")

	a.Run()

	defer a.RequestHandler.CommandScheduler.Shutdown()
}
