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
	//DefaultConfigDir = ".recmd"

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
	Secret         Secret
	Footprint      Footprint
	DmnLogFile     LogFile
	History        HistoryFile
}

func (a *App) InitializeProd() {

	footprint := Footprint{}
	footprint.DefaultFootprint()

	a.Footprint.Set(&footprint)

	// Set the secret file
	a.Secret.Set(footprint.confDirPath)
	a.Secret.WriteSecretToFile()

	// Set the log file
	a.DmnLogFile.Set(footprint.logDirPath)
	a.DmnLogFile.Create()

	// Set the history file
	a.History.Set(footprint.confDirPath)
	a.History.WriteHistoryToFile()

	a.DmnLogFile.Log.Printf("Initializing...")

	// Server code
	a.Server = http.Server{Addr: DefaultServerPort, Handler: nil}

	a.Router = mux.NewRouter()

	a.InitializeRoutes()

	a.RequestHandler.CommandScheduler.CreateScheduler()
	go a.RequestHandler.CommandScheduler.RunScheduler()
	go a.RequestHandler.CommandScheduler.QueuedCommandsCleanup()
}

// InitalizeTest deletes and recreates the testdata sandbox directory for testing.
// Global variables used by the test are also initialized.
func (a *App) InitalizeTest() error {

	fmt.Println("Intializing testdata dir...")

	footprint := Footprint{}
	footprint.TestFootprint()

	a.Footprint.Set(&footprint)

	// Set the secret file
	a.Secret.Set(footprint.confDirPath)
	err := a.Secret.WriteSecretToFile()
	if err != nil {
		return err
	}
	if a.Secret.GetSecret() == "" {
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
			a.DmnLogFile.Log.Fatalf("Error, unable to create ~/.recmd: %v\n", err)
		}
	} else if !fileInfo.IsDir() {
		a.DmnLogFile.Log.Fatalf("Error, ~/.recmd is not a directory")
	}
}

// CreateLogs creates a new log file
func (a *App) CreateLogs() {

	workingDirectory, err := os.Getwd()

	if err != nil {
		a.DmnLogFile.Log.Fatalf("Error when creating logs: %v\n", err)
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
		a.DmnLogFile.Log.Fatalf("error opening file: %v", err)
	}

	a.DmnLogFile.Log = log.New(f, "", log.LstdFlags|log.Lshortfile)

}

// InitializeRoutes initializes the routes for this application
func (a *App) InitializeRoutes() {
	a.Router.HandleFunc("/secret/{secret}/delete/cmdHash/{cmdHash}", a.HandleDelete)
	a.Router.HandleFunc("/secret/{secret}/add/command/{command}/description/{description}/workingDirectory/{workingDirectory}", a.HandleAdd)
	a.Router.HandleFunc("/secret/{secret}/select/cmdHash/{cmdHash}", a.HandleSelect)
	a.Router.HandleFunc("/secret/{secret}/search/description/{description}", a.HandleSearch)
	a.Router.HandleFunc("/secret/{secret}/run/cmdHash/{cmdHash}", a.HandleRun)
	a.Router.HandleFunc("/secret/{secret}/list", a.HandleList)
	a.Router.HandleFunc("/secret/{secret}/queue", a.HandleQueue)

	http.Handle("/", a.Router)
}

// Run runs the application
func (a *App) Run() {
	a.DmnLogFile.Log.Printf("Starting server on %v\n", DefaultServerPort)

	//http.ListenAndServe(DefaultServerPort, nil)
	if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// Shutdown shuts down the http server
func (a *App) Shutdown() {
	a.DmnLogFile.Log.Printf("Shutting down server")
	a.Server.Shutdown(context.Background())
}

// Execute is a convenience function that runs the program and quits if there is a signal.
func Execute() {

	var a App

	a.InitializeProd()

	a.DmnLogFile.Log.Printf("Starting up...")

	go func() {
		a.Run()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	a.Shutdown()
}
