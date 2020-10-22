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
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	dmn "github.com/tarof429/recmd-dmn/dmn"
)

// ScheduledCommand represents a dmn.Command that is scheduled to run
type ScheduledCommand struct {
	dmn.Command
	Coutput    string    `json:"coutput"`
	ExitStatus int       `json:"exitStatus"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
}

const (
	// Port that this program listens to
	serverPort = ":8999"

	// Directory containing configuration and dmn.Command history
	recmdDir = ".recmd"

	// The dmn.Command history file
	recmdHistoryFile = "recmd_history.json"
)

// Global variables
var (
	recmdDirPath        string
	recmdSecretFilePath string
	secretData          string
	cmdHistoryFilePath  string
	secret              *dmn.Secret
)

// ReadCmdHistoryFile reads historyFile and generates a list of dmn.Command structs
func ReadCmdHistoryFile() ([]dmn.Command, error) {

	var (
		historyData []byte        // Data representing our history file
		cmds        []dmn.Command // List of dmn.Commands produced after unmarshalling historyData
		err         error         // Any errors we might encounter
	)

	// Read the history file into historyData
	historyData, err = ioutil.ReadFile(cmdHistoryFilePath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred while reading historyfile: %v\n", err)
		return cmds, err
	}

	// Unmarshall historyData into a list of dmn.Commands
	err = json.Unmarshal(historyData, &cmds)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while unmarshalling: %v\n", err)
	}

	return cmds, err

}

// SelectCmd returns a dmn.Command
func SelectCmd(dir string, value string) (dmn.Command, error) {

	log.Println("Selecting " + value)

	cmds, error := ReadCmdHistoryFile()

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

	cmds, error := ReadCmdHistoryFile()

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

// DeleteCmd deletes a dmn.Command. It's best to pass in the dmn.CommandHash
// because dmn.Commands may look similar.
func DeleteCmd(value string) ([]dmn.Command, error) {

	log.Println("Deleting " + value)

	ret := []dmn.Command{}

	cmds, error := ReadCmdHistoryFile()

	if error != nil {
		return ret, error
	}

	foundIndex := -1

	for index, cmd := range cmds {
		if strings.Index(cmd.CmdHash, value) == 0 {
			foundIndex = index
			break
		}
	}

	if foundIndex != -1 {
		ret = append(ret, cmds[foundIndex])

		//fmt.Println("Found dmn.Command. Found index was " + strconv.Itoa(foundIndex))
		// We may want to do more investigation to know why this works...
		cmds = append(cmds[:foundIndex], cmds[foundIndex+1:]...)

		// Return whether we are able to overwrite the history file
		OverwriteCmdHistoryFile(cmds)
	}

	return ret, nil
}

// OverwriteCmdHistoryFile overwrites the history file with []dmn.Command passed in as a parameter
func OverwriteCmdHistoryFile(cmds []dmn.Command) bool {

	mode := int(0644)

	updatedData, _ := json.MarshalIndent(cmds, "", "\t")

	error := ioutil.WriteFile(cmdHistoryFilePath, updatedData, os.FileMode(mode))

	return error == nil
}

// CreateCmdHistoryFile creates an empty history file
func CreateCmdHistoryFile() bool {

	// Check if the file does not exist. If not, then create it and add our first dmn.Command to it.
	f, err := os.Open(cmdHistoryFilePath)

	// Immediately close the file since we plan to write to it
	defer f.Close()

	// Check if the file doesn't exist and if so, then write it.
	if err != nil {

		mode := int(0644)

		error := ioutil.WriteFile(cmdHistoryFilePath, []byte(nil), os.FileMode(mode))

		return error == nil
	}
	return true
}

// UpdateCommandDuration updates a dmn.Command with the same hash in the history file
func UpdateCommandDuration(cmd dmn.Command, duration time.Duration) bool {

	// Check if the file does not exist. If not, then create it and add our first dmn.Command to it.
	f, err := os.Open(cmdHistoryFilePath)

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

		error := ioutil.WriteFile(cmdHistoryFilePath, updatedData, os.FileMode(mode))

		return error == nil
	}

	// Update the dmn.Command in the history file

	// The array of dmn.Commands
	var cmds []dmn.Command

	// Read history file
	data, err := ioutil.ReadFile(cmdHistoryFilePath)

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

	error := ioutil.WriteFile(cmdHistoryFilePath, updatedData, os.FileMode(mode))

	return error == nil
}

// WriteCmdHistoryFile writes a dmn.Command to the history file
func WriteCmdHistoryFile(cmd dmn.Command) bool {

	// Check if the file does not exist. If not, then create it and add our first dmn.Command to it.
	f, err := os.Open(cmdHistoryFilePath)

	// Immediately close the file since we plan to write to it
	f.Close()

	// Check if the file doesn't exist and if so, then write it.
	if err != nil {
		// The array of dmn.Commands
		var cmds []dmn.Command

		cmds = append(cmds, cmd)

		mode := int(0644)

		updatedData, _ := json.MarshalIndent(cmds, "", "\t")

		error := ioutil.WriteFile(cmdHistoryFilePath, updatedData, os.FileMode(mode))

		return error == nil
	}

	// Update the dmn.Command in the history file

	// The array of dmn.Commands
	var cmds []dmn.Command

	// Read history file
	data, err := ioutil.ReadFile(cmdHistoryFilePath)

	// An error occured while reading historyFile.
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return false
	}

	// No data in file, write our dmn.Command to it
	if len(data) == 0 {
		cmds = append(cmds, cmd)
		updatedData, _ := json.MarshalIndent(cmds, "", "\t")
		mode := int(0644)
		error := ioutil.WriteFile(cmdHistoryFilePath, updatedData, os.FileMode(mode))
		return error == nil
	}
	if err := json.Unmarshal(data, &cmds); err != nil {
		fmt.Fprintf(os.Stderr, "JSON unmarshalling failed: %s\n", err)
		return false
	}

	// Check if the dmn.Command hash alaready exists, and prevent the user from adding the same dmn.Command
	for _, c := range cmds {
		if c.CmdHash == cmd.CmdHash {
			// c.Duration = cmd.Duration
			fmt.Fprintf(os.Stderr, "dmn.Command hash already exists: %s\n", cmd.CmdString)
			//break
			return false
		}
	}

	cmds = append(cmds, cmd)

	// Convert the struct to JSON
	updatedData, updatedDataErr := json.MarshalIndent(cmds, "", "\t")

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", updatedDataErr)
	}

	mode := int(0644)

	error := ioutil.WriteFile(cmdHistoryFilePath, updatedData, os.FileMode(mode))

	return error == nil

}

// NewCommand creates a new dmn.Command struct and populates the fields
func NewCommand(cmdString string, cmdComment string) dmn.Command {

	formattedHash := func() string {
		h := sha1.New()
		h.Write([]byte(cmdString))
		return fmt.Sprintf("%.15x", h.Sum(nil))
	}()

	cmd := dmn.Command{CmdHash: formattedHash,
		CmdString:   strings.Trim(cmdString, ""),
		Description: strings.Trim(cmdComment, ""),
		Duration:    -1,
	}

	return cmd
}

// ScheduleCommand runs a dmn.Command based on a function passed in as the second parameter.
// This gives the ability to run dmn.Commands in multiple ways; for example, as a "mock" dmn.Command
// (RunMockdmn.Command) or a shell script dmn.Command (RunShellScriptdmn.Command).
func ScheduleCommand(cmd dmn.Command, f func(*ScheduledCommand, chan int)) ScheduledCommand {
	var sc ScheduledCommand

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

// RunMockCommand runs a mock dmn.Command
func RunMockCommand(sc *ScheduledCommand, c chan int) {
	time.Sleep(1 * time.Second)
	sc.ExitStatus = 99
	sc.Coutput = "Mock stdout message"
	c <- sc.ExitStatus
}

// RunShellScriptCommand runs a dmn.Command written to a temporary file
func RunShellScriptCommand(sc *ScheduledCommand, c chan int) {

	tempFile, err := ioutil.TempFile(os.TempDir(), "recmd-")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to create temp file: %d\n", err)
	}

	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("#!/bin/sh\n\n" + sc.CmdString)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Errror: unable to write script to temp file: : %s\n", err)
	}

	cmd := exec.Command("sh", tempFile.Name())

	// We may want to make this configurable in the future.
	// For now, all dmn.Commands will be run from the user's home directory
	cmd.Dir, err = os.UserHomeDir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to obtain home directory: %s\n", err)
	}

	//out, err := cmd.Output()

	combinedOutput, combinedOutputErr := cmd.CombinedOutput()

	// fmt.Fprintf(os.Stdout, "\nError: %s error 2: %v\n", string(combinedOutput), err2)

	if combinedOutputErr != nil {
		sc.ExitStatus = -1
	}

	sc.Coutput = string(combinedOutput)

	c <- sc.ExitStatus
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

	ret, err := ReadCmdHistoryFile()

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
		log.Println("Unable to select dmn.Command")
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
		log.Println("Unable to select dmn.Command")
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

// addHandler adds a dmn.Command
func addHandler(w http.ResponseWriter, r *http.Request) {

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
		out, _ := json.Marshal("false")
		io.WriteString(w, string(out))
		return
	}

	// Select the dmn.Command, otherwise, if the dmn.Command hash cannot be found, return error 400
	testCmd := new(dmn.Command)
	testCmd.Set(variables.Command, variables.Description)
	//testCmd := Newdmn.Command(variables.dmn.Command, variables.Description)

	if WriteCmdHistoryFile(*testCmd) != true {
		w.WriteHeader(http.StatusBadRequest)
		out, _ := json.Marshal("false")
		io.WriteString(w, string(out))
		return
	}

	w.WriteHeader(http.StatusOK)
	out, _ := json.Marshal("true")
	io.WriteString(w, string(out))
}

// deleteHandler deletess a dmn.Command
func deleteHandler(w http.ResponseWriter, r *http.Request) {

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
	selectedCmd, err := DeleteCmd(variables.CmdHash)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
		log.Println("Unable to select dmn.Command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if selectedCmd.CmdHash == "" {
		log.Println("Invalid hash")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("Scheduling dmn.Command")

	sc := ScheduleCommand(selectedCmd, RunShellScriptCommand)

	log.Println("dmn.Command completed")

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

// InitTool initializes the tool
func InitTool() {

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

	// Load the dmn.Command history file path. We don't need to read it yet.
	cmdHistoryFilePath = filepath.Join(recmdDirPath, recmdHistoryFile)

	//recmdHistoryFile := filepath.Join(recmdDirPath, recmdHistoryFile)

	_, statErr = os.Stat(cmdHistoryFilePath)

	if os.IsNotExist(statErr) {
		CreateCmdHistoryFile()
	}
	// Create the secrets file containing the secret
	secret = new(dmn.Secret)
	secret.CreateSecret()
	secret.SetPathToSecretsFile(recmdDirPath)
	err = secret.WriteSecretToFile()

	if err != nil {
		log.Fatalf("Error, unable to create secrets file %v\n", err)
		return
	}
}

func main() {
	InitTool()
	r := mux.NewRouter()

	r.HandleFunc("/secret/{secret}/delete/cmdHash/{cmdHash}", deleteHandler)
	r.HandleFunc("/secret/{secret}/select/cmdHash/{cmdHash}", selectHandler)
	r.HandleFunc("/secret/{secret}/search/description/{description}", searchHandler)
	r.HandleFunc("/secret/{secret}/run/cmdHash/{cmdHash}", runHandler)
	r.HandleFunc("/secret/{secret}/list", listHandler)
	r.HandleFunc("/secret/{secret}/add/dmn.Command/{dmn.Command}/description/{description}", addHandler)

	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(serverPort, nil))

}
