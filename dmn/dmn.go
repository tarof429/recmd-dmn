package dmn

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/gorilla/mux"
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
	Server         http.Server
	RequestHandler RequestHandler
	DmnSecret      Secret
	Footprint      Footprint
	DmnLogFile     LogFile
	History        HistoryFile
}

// Initialize initializes the application
func (a *App) Initialize(configPath string) {

	a.Server = http.Server{Addr: DefaultServerPort, Handler: nil}

	a.CreateLogs()

	a.Router = mux.NewRouter()

	a.InitializeConfigPath(configPath)

	//secret := a.CreateSecret()

	//historyFile := a.CreateHistoryFile()

	//a.RequestHandler.Set(secret, historyFile)

	a.InitializeRoutes()

	a.RequestHandler.CommandScheduler.CreateScheduler()
	go a.RequestHandler.CommandScheduler.RunScheduler()
	go a.RequestHandler.CommandScheduler.QueuedCommandsCleanup()
}

// InitHandlerTest deletes and recreates the testdata sandbox directory for testing.
// Global variables used by the test are also initialized.
func (a *App) InitHandlerTest() error {

	fmt.Println("Intializing testdata dir...")

	footprint := Footprint{}
	footprint.TestFootprint()

	a.Footprint.Set(&footprint)

	// Set the secret file
	a.DmnSecret.Set(footprint.confDirPath)
	err := a.DmnSecret.WriteSecretToFile()
	if err != nil {
		return err
	}
	if a.DmnSecret.GetSecret() == "" {
		return err
	}

	// Set the log file
	a.DmnLogFile.Set(footprint.logDirPath)
	a.DmnLogFile.Create()

	// Set the history file
	a.History.Set(footprint.confDirPath)
	a.History.Remove()
	err = a.History.WriteHistoryToFile()
	if err != nil {
		return err
	}

	a.RequestHandler.Set(a.DmnSecret,
		a.History, a.DmnLogFile.Log)

	return nil

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

	workingDirectory, err := os.Getwd()

	if err != nil {
		log.Fatalf("Error when creasting logs: %v\n", err)
	}

	// here
	logsDir := filepath.Join(workingDirectory, "logs")

	var logFile string

	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		fmt.Println("Using default log file since logs directory does not exist")
		logFile = DefaultLogFile
	} else {
		fmt.Println("Using logs file in logs directory")
		logFile = filepath.Join(logsDir, DefaultLogFile)
	}

	f, err := os.Create(logFile)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	a.RequestHandler.Log = log.New(f, "", log.LstdFlags|log.Lshortfile)

}

// CreateSecret creates the secret whenever the application starts
func (a *App) CreateSecret() Secret {

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
func (a *App) GetSecret() Secret {
	return a.DmnSecret
}

// CreateHistoryFile initializes the historyFile file
func (a *App) CreateHistoryFile() HistoryFile {

	var historyFile HistoryFile

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
	a.Router.HandleFunc("/secret/{secret}/delete/cmdHash/{cmdHash}", a.HandleDelete)
	a.Router.HandleFunc("/secret/{secret}/add/command/{command}/description/{description}/workingDirectory/{workingDirectory}", a.HandleAdd)
	a.Router.HandleFunc("/secret/{secret}/select/cmdHash/{cmdHash}", a.HandleSelect)
	a.Router.HandleFunc("/secret/{secret}/search/description/{description}", a.HandleSearch)
	a.Router.HandleFunc("/secret/{secret}/run/cmdHash/{cmdHash}", a.HandleRun)
	a.Router.HandleFunc("/secret/{secret}/list", a.HandleList)
	a.Router.HandleFunc("/secret/{secret}/status", a.HandleStatus)

	http.Handle("/", a.Router)
}

// Run runs the application
func (a *App) Run() {
	log.Printf("Starting server on %v\n", DefaultServerPort)

	//http.ListenAndServe(DefaultServerPort, nil)
	if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// Shutdown shuts down the http server
func (a *App) Shutdown() {
	log.Printf("Shutting down server")
	a.Server.Shutdown(context.Background())
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

// Execute is a convenience function that runs the program and quits if there is a signal.
func Execute() {
	a := App{}

	configPath := GetDefaultConfigPath()

	a.Initialize(configPath)

	a.RequestHandler.Log.Printf("Starting up!")

	go func() {
		a.Run()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	a.Shutdown()
}
