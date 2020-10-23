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
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	secretData    string
	secret        dmn.Secret
	history       dmn.HistoryFile
	deleteHandler dmn.DeleteHandler
	addHandler    dmn.AddHandler
)

// SelectCmd returns a dmn.Command
func SelectCmd(dir string, value string) (dmn.Command, error) {

	log.Println("Selecting " + value)

	cmds, error := history.ReadCmdHistoryFile()

	if error != nil {
		return dmn.Command{}, error
	}

	for _, cmd := range cmds {

		if strings.Index(cmd.CmdHash, value) == 0 {
			return cmd, nil
		}
	}

	return dmn.Command{}, nil
}

// SearchCmd returns a dmn.Command by name
func SearchCmd(value string) ([]dmn.Command, error) {

	log.Println("Searching " + value)

	cmds, error := history.ReadCmdHistoryFile()

	ret := []dmn.Command{}

	if error != nil {
		return []dmn.Command{}, error
	}

	for _, cmd := range cmds {

		// Use lower case for evaluation
		comment := strings.ToLower(cmd.Description)

		if strings.Contains(comment, value) {
			ret = append(ret, cmd)
		}
	}

	return ret, nil
}

// UpdateCommandDuration updates a dmn.Command with the same hash in the history file
func UpdateCommandDuration(cmd dmn.Command, duration time.Duration) bool {

	// Check if the file does not exist. If not, then create it and add our first dmn.Command to it.
	f, err := os.Open(history.Path)

	// Immediately close the file since we plan to write to it
	f.Close()

	// Check if the file doesn't exist and if so, then write it.
	if err != nil {

		// The array of dmn.Commands
		var cmds []dmn.Command

		// Set the duration
		cmd.Duration = duration

		cmds = append(cmds, cmd)

		mode := int(0644)

		updatedData, _ := json.MarshalIndent(cmds, "", "\t")

		error := ioutil.WriteFile(history.Path, updatedData, os.FileMode(mode))

		return error == nil
	}

	// Update the dmn.Command in the history file

	// The array of dmn.Commands
	var cmds []dmn.Command

	// Read history file
	data, err := ioutil.ReadFile(history.Path)

	// An error occured while reading historyFile.
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return false
	}

	if err := json.Unmarshal(data, &cmds); err != nil {
		fmt.Fprintf(os.Stderr, "JSON unmarshalling failed: %s\n", err)
		return false
	}

	//fmt.Println("Updating duration")

	var found bool
	var foundIndex int

	// Update the duration for the dmn.Command
	for index, c := range cmds {
		if c.CmdHash == cmd.CmdHash {
			//fmt.Println("Found dmn.Command")
			foundIndex = index
			found = true
			//c.Duration = cmd.Duration
			//fmt.Fprintf(os.Stderr, "dmn.Command hash already exists: %s\n", cmd.CmdString)
			break
			//return false
		}
	}

	if found == true {
		cmds[foundIndex].Duration = duration
		//fmt.Println(cmds[foundIndex])
	}

	// Convert the struct to JSON
	updatedData, updatedDataErr := json.MarshalIndent(cmds, "", "\t")

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", updatedDataErr)
	}

	mode := int(0644)

	error := ioutil.WriteFile(history.Path, updatedData, os.FileMode(mode))

	return error == nil
}

// ScheduleCommand runs a dmn.Command based on a function passed in as the second parameter.
// This gives the ability to run dmn.Commands in multiple ways; for example, as a "mock" dmn.Command
// (RunMockdmn.Command) or a shell script dmn.Command (RunShellScriptdmn.Command).
func ScheduleCommand(cmd dmn.Command, f func(*dmn.ScheduledCommand, chan int)) dmn.ScheduledCommand {
	var sc dmn.ScheduledCommand

	sc.CmdHash = cmd.CmdHash
	sc.CmdString = cmd.CmdString
	sc.Description = cmd.Description
	sc.Duration = -1

	// Create a channel to hold exit status
	c := make(chan int)

	// Set the start time
	sc.StartTime = time.Now()

	// Run the dmn.Command in a goroutine
	go f(&sc, c)

	// Receive the exit status of the dmn.Command
	status := <-c

	now := time.Now()

	// Set end time after we receive from the channel
	sc.EndTime = now

	// Calculate the duration and store it
	sc.Duration = now.Sub(sc.StartTime)

	// The main reason why this code exists is to use the value received from the channel.
	if status != 0 {
		fmt.Fprintf(os.Stderr, "\nError: dmn.Command failed.\n")
	}

	return sc
}

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

// searchHandler selects a dmn.Command
func searchHandler(w http.ResponseWriter, r *http.Request) {

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

	// Select the dmn.Command, otherwise, if the dmn.Command hash cannot be found, return error 400
	selectedCmds, cerr := SearchCmd(variables.Description)

	if cerr != nil {
		log.Println("Unable to select Command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	out, err := json.Marshal(selectedCmds)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	io.WriteString(w, string(out))
}

// selectHandler selects a dmn.Command
func selectHandler(w http.ResponseWriter, r *http.Request) {

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

	// Select the dmn.Command, otherwise, if the dmn.Command hash cannot be found, return error 400
	selectedCmd, cerr := SelectCmd(recmdDirPath, variables.CmdHash)

	if cerr != nil {
		log.Println("Unable to select Command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if selectedCmd.CmdHash == "" {
		log.Println("Invalid hash")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	out, err := json.Marshal(selectedCmd)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	io.WriteString(w, string(out))
}

// runHandler runs a dmn.Command
func runHandler(w http.ResponseWriter, r *http.Request) {

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

	// Select the dmn.Command, otherwise, if the dmn.Command hash cannot be found, return error 400
	selectedCmd, cerr := SelectCmd(recmdDirPath, variables.CmdHash)

	if cerr != nil {
		log.Println("Unable to select Command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if selectedCmd.CmdHash == "" {
		log.Println("Invalid hash")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("Scheduling Command")

	sc := ScheduleCommand(selectedCmd, dmn.RunShellScriptCommand)

	log.Println("Command completed")

	if len(sc.Coutput) != 0 {
		// fmt.Println(sc.Coutput)
		w.WriteHeader(http.StatusOK)

		out, err := json.Marshal(sc)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		io.WriteString(w, string(out))
	}

	ret := UpdateCommandDuration(selectedCmd, sc.Duration)

	if ret != true {
		w.WriteHeader(http.StatusBadRequest)
	}

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

	deleteHandler.Set(secret, history)
	r.HandleFunc("/secret/{secret}/delete/cmdHash/{cmdHash}", deleteHandler.Handle)

	addHandler.Set(secret, history)
	r.HandleFunc("/secret/{secret}/add/command/{command}/description/{description}", addHandler.Handle)

	r.HandleFunc("/secret/{secret}/select/cmdHash/{cmdHash}", selectHandler)
	r.HandleFunc("/secret/{secret}/search/description/{description}", searchHandler)
	r.HandleFunc("/secret/{secret}/run/cmdHash/{cmdHash}", runHandler)
	r.HandleFunc("/secret/{secret}/list", listHandler)

	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(serverPort, nil))

}
