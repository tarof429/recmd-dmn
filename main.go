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
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	dmn "github.com/tarof429/recmd-dmn/dmn"
)

const (
	// Port that this program listens to
	serverPort = ":8999"

	// Directory containing configuration and dmn.Command history
	recmdDir = ".recmd"
)

// Global variables
var (
	recmdDirPath string
	//recmdSecretFilePath string
	secretData string
	secret     dmn.Secret
	history    dmn.HistoryFile

	requestHandler dmn.RequestHandler
)

// listHandler lists dmn.Commands
func listHandler(w http.ResponseWriter, r *http.Request) {

	// Get variables from the request
	vars := mux.Vars(r)
	var variables dmn.RequestVariable
	err := variables.GetVariablesFromRequestVars(vars)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the secret we passed in is valid, otherwise, return error 400
	if !secret.Valid(variables.Secret) {
		log.Println("Bad secret!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ret, err := history.ReadCmdHistoryFile()

	if err != nil {
		log.Println("Unable to read history file")
	}

	w.WriteHeader(http.StatusOK)

	out, err := json.Marshal(ret)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	io.WriteString(w, string(out))
}

// initTool initializes the tool
func initTool() {

	// Create ~/.recmd if it doesn't exist
	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatalf("Error, unable to obtain home directory path %v\n", err)
	}

	recmdDirPath = filepath.Join(homeDir, recmdDir)
	fileInfo, statErr := os.Stat(recmdDirPath)

	if os.IsNotExist((statErr)) {
		mode := int(0755)

		err = os.Mkdir(recmdDirPath, os.FileMode(mode))

		if err != nil {
			log.Fatalf("Error, unable to create ~/.recmd: %v\n", err)
		}
	} else if !fileInfo.IsDir() {
		log.Fatalf("Error, ~/.recmd is not a directory")
	}

	// Every time this program starts, create a new secret
	secret.Set(recmdDirPath)
	err = secret.WriteSecretToFile()
	if err != nil {
		log.Fatalf("Error, unable to create secrets file %v\n", err)
		return
	}
	if secret.GetSecret() == "" {
		log.Fatalf("Error, secret was an empty string")
		return
	}

	// Load the history file. If it doesn't exist, create it.
	history.Set(recmdDirPath)
	_, statErr = os.Stat(history.Path)
	if os.IsNotExist(statErr) {
		err = history.WriteHistoryToFile()
		if err != nil {
			log.Fatalf("Error, unable to create history file")
			return
		}
	}

}

func main() {
	initTool()

	r := mux.NewRouter()

	requestHandler.Set(secret, history)

	r.HandleFunc("/secret/{secret}/delete/cmdHash/{cmdHash}", requestHandler.HandleDelete)
	r.HandleFunc("/secret/{secret}/add/command/{command}/description/{description}", requestHandler.HandleAdd)
	r.HandleFunc("/secret/{secret}/select/cmdHash/{cmdHash}", requestHandler.HandleSelect)
	r.HandleFunc("/secret/{secret}/search/description/{description}", requestHandler.HandleSelect)
	r.HandleFunc("/secret/{secret}/run/cmdHash/{cmdHash}", requestHandler.HandleRun)

	r.HandleFunc("/secret/{secret}/list", listHandler)

	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(serverPort, nil))

}
