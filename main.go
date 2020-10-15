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
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Command represents a command and optionally a description to document what the command does
type Command struct {
	CmdHash     string        `json:"commandHash"`
	CmdString   string        `json:"commandString"`
	Description string        `json:"description"`
	Duration    time.Duration `json:"duration"`
}

// ScheduledCommand represents a command that is scheduled to run
type ScheduledCommand struct {
	Command
	Coutput    string    `json:"coutput"`
	ExitStatus int       `json:"exitStatus"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
}

const (
	// Directory containing configuration and command history
	recmdDir = ".recmd"

	// The secret file
	recmdSecretFile = "recmd_secret"

	// The command history file
	recmdHistoryFile = "recmd_history.json"

	// List of characters in our secret
	secretCharSet = "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP123456789"

	// Length of scret string
	secretLength = 40
)

// Global variables
var (
	recmdDirPath        string
	recmdSecretFilePath string
	secretData          string
	cmdHistoryFilePath  string
)

// ReadCmdHistoryFile reads historyFile and generates a list of Command structs
func ReadCmdHistoryFile() ([]Command, error) {

	var (
		historyData []byte    // Data representing our history file
		cmds        []Command // List of commands produced after unmarshalling historyData
		err         error     // Any errors we might encounter
	)

	// Read the history file into historyData
	historyData, err = ioutil.ReadFile(cmdHistoryFilePath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred while reading historyfile: %v\n", err)
		return cmds, err
	}

	// Unmarshall historyData into a list of commands
	err = json.Unmarshal(historyData, &cmds)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while unmarshalling: %v\n", err)
	}

	return cmds, err

}

// SelectCmd returns a command
func SelectCmd(dir string, value string) (Command, error) {

	cmds, error := ReadCmdHistoryFile()

	if error != nil {
		return Command{}, error
	}

	for _, cmd := range cmds {

		if strings.Index(cmd.CmdHash, value) == 0 {
			return cmd, nil
		}
	}

	return Command{}, nil
}

// SearchCmd returns a command by name
func SearchCmd(value string) ([]Command, error) {

	cmds, error := ReadCmdHistoryFile()

	ret := []Command{}

	if error != nil {
		return []Command{}, error
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

// DeleteCmd deletes a command. It's best to pass in the commandHash
// because commands may look similar.
func DeleteCmd(value string) int {

	cmds, error := ReadCmdHistoryFile()

	if error != nil {
		return -1
	}

	foundIndex := -1

	for index, cmd := range cmds {
		if strings.Index(cmd.CmdHash, value) == 0 {
			foundIndex = index
			break
		}
	}

	if foundIndex != -1 {
		//fmt.Println("Found command. Found index was " + strconv.Itoa(foundIndex))
		// We may want to do more investigation to know why this works...
		cmds = append(cmds[:foundIndex], cmds[foundIndex+1:]...)

		// Return whether we are able to overwrite the history file
		OverwriteCmdHistoryFile(cmds)
	}

	return foundIndex
}

// OverwriteCmdHistoryFile overwrites the history file with []Command passed in as a parameter
func OverwriteCmdHistoryFile(cmds []Command) bool {

	mode := int(0644)

	updatedData, _ := json.MarshalIndent(cmds, "", "\t")

	error := ioutil.WriteFile(cmdHistoryFilePath, updatedData, os.FileMode(mode))

	return error == nil
}

// CreateCmdHistoryFile creates an empty history file
func CreateCmdHistoryFile() bool {

	// Check if the file does not exist. If not, then create it and add our first command to it.
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

// UpdateCommandDuration updates a command with the same hash in the history file
func UpdateCommandDuration(cmd Command, duration time.Duration) bool {

	// Check if the file does not exist. If not, then create it and add our first command to it.
	f, err := os.Open(cmdHistoryFilePath)

	// Immediately close the file since we plan to write to it
	f.Close()

	// Check if the file doesn't exist and if so, then write it.
	if err != nil {

		// The array of commands
		var cmds []Command

		// Set the duration
		cmd.Duration = duration

		cmds = append(cmds, cmd)

		mode := int(0644)

		updatedData, _ := json.MarshalIndent(cmds, "", "\t")

		error := ioutil.WriteFile(cmdHistoryFilePath, updatedData, os.FileMode(mode))

		return error == nil
	}

	// Update the command in the history file

	// The array of commands
	var cmds []Command

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

	// Update the duration for the command
	for index, c := range cmds {
		if c.CmdHash == cmd.CmdHash {
			//fmt.Println("Found command")
			foundIndex = index
			found = true
			//c.Duration = cmd.Duration
			//fmt.Fprintf(os.Stderr, "Command hash already exists: %s\n", cmd.CmdString)
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

// WriteCmdHistoryFile writes a command to the history file
func WriteCmdHistoryFile(cmd Command) bool {

	// Check if the file does not exist. If not, then create it and add our first command to it.
	f, err := os.Open(cmdHistoryFilePath)

	// Immediately close the file since we plan to write to it
	f.Close()

	// Check if the file doesn't exist and if so, then write it.
	if err != nil {
		// The array of commands
		var cmds []Command

		cmds = append(cmds, cmd)

		mode := int(0644)

		updatedData, _ := json.MarshalIndent(cmds, "", "\t")

		error := ioutil.WriteFile(cmdHistoryFilePath, updatedData, os.FileMode(mode))

		return error == nil
	}

	// Update the command in the history file

	// The array of commands
	var cmds []Command

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

	// Check if the command hash alaready exists, and prevent the user from adding the same command
	for _, c := range cmds {
		if c.CmdHash == cmd.CmdHash {
			// c.Duration = cmd.Duration
			fmt.Fprintf(os.Stderr, "Command hash already exists: %s\n", cmd.CmdString)
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

// NewCommand creates a new Command struct and populates the fields
func NewCommand(cmdString string, cmdComment string) Command {

	formattedHash := func() string {
		h := sha1.New()
		h.Write([]byte(cmdString))
		return fmt.Sprintf("%.15x", h.Sum(nil))
	}()

	cmd := Command{formattedHash,
		strings.Trim(cmdString, ""),
		strings.Trim(cmdComment, ""),
		-1}

	return cmd
}

// ScheduleCommand runs a Command based on a function passed in as the second parameter.
// This gives the ability to run Commands in multiple ways; for example, as a "mock" command
// (RunMockCommand) or a shell script command (RunShellScriptCommand).
func ScheduleCommand(cmd Command, f func(*ScheduledCommand, chan int)) ScheduledCommand {
	var sc ScheduledCommand

	sc.CmdHash = cmd.CmdHash
	sc.CmdString = cmd.CmdString
	sc.Description = cmd.Description
	sc.Duration = -1

	// Create a channel to hold exit status
	c := make(chan int)

	// Set the start time
	sc.StartTime = time.Now()

	// Run the command in a goroutine
	go f(&sc, c)

	// Receive the exit status of the command
	status := <-c

	now := time.Now()

	// Set end time after we receive from the channel
	sc.EndTime = now

	// Calculate the duration and store it
	sc.Duration = now.Sub(sc.StartTime)

	// The main reason why this code exists is to use the value received from the channel.
	if status != 0 {
		fmt.Fprintf(os.Stderr, "\nError: command failed.\n")
	}

	return sc
}

// RunMockCommand runs a mock command
func RunMockCommand(sc *ScheduledCommand, c chan int) {
	time.Sleep(1 * time.Second)
	sc.ExitStatus = 99
	sc.Coutput = "Mock stdout message"
	c <- sc.ExitStatus
}

// RunShellScriptCommand runs a command written to a temporary file
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
	// For now, all commands will be run from the user's home directory
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

// CmdHandler handles a request for http://localhost:8090/cmd/<hash>
func CmdHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// Get the requested command hash
	vars := mux.Vars(r)
	cmdHash := vars["hash"]
	secret := vars["secret"]

	// Check if the secret we passed in is valid, otherwise, return error 400
	if secret != GetSecret() {
		log.Println("Bad secret!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Select the command, otherwise, if the command hash cannot be found, return error 400
	selectedCmd, cerr := SelectCmd(recmdDirPath, cmdHash)

	if cerr != nil {
		log.Println("Unable to select command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if selectedCmd.CmdHash == "" {
		log.Println("Invalid hash")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("Scheduling command")

	// Schedule the command. This code may need to be cleaned up since we set the header twice.
	sc := ScheduleCommand(selectedCmd, RunShellScriptCommand)

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

// CreateSecret create a secret string. This must be passed to the service to call the method successfully.
func CreateSecret(charSet string) string {

	rand.Seed(time.Now().Unix())

	var s string

	for i := 0; i < secretLength; i++ {
		random := rand.Intn(len(charSet))
		randomChar := charSet[random]
		s += string(randomChar)
	}
	return s
}

// CreateSecretsFile creates the file containing the secret
func CreateSecretsFile(secret string) error {

	file, err := os.Create(recmdSecretFilePath)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.WriteString(file, secret)

	if err != nil {
		return err
	}

	return file.Sync()
}

// GetSecret gets the secret from the file system
func GetSecret() string {
	secretData, err := ioutil.ReadFile(recmdSecretFilePath)

	if err != nil {
		log.Fatalf("Error, unable to read secret from file %v\n", err)
	}

	if len(secretData) != secretLength {
		log.Fatalf("Error, invalid secret length %v\n", err)
	}

	return string(secretData)
}

// Initialize the tool
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

	// Create the secrets file. Will be recreated each time this
	// function is called.
	recmdSecretFilePath = filepath.Join(recmdDirPath, recmdSecretFile)

	secret := CreateSecret(secretCharSet)

	err = CreateSecretsFile(secret)

	if err != nil {
		log.Fatalf("Error, unable to create secrets file %v\n", err)
		return
	}

	// Load the command history file path. We don't need to read it yet.
	cmdHistoryFilePath = filepath.Join(recmdDirPath, recmdHistoryFile)
}

func main() {
	initTool()
	r := mux.NewRouter()
	r.HandleFunc("/secret/{secret}/hash/{hash}", CmdHandler)
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8090", nil))

	// // Maybe the following should be written as a test!
	// cmdHash := "37fa265330ad83eaa879efb1e2db63"

	// selectedCmd, cerr := SelectCmd(recmdDirPath, cmdHash)

	// if cerr != nil {
	// 	fmt.Fprintf(os.Stderr, "Error: unable to read history file: %s\n", cerr)
	// 	return
	// }

	// sc := ScheduleCommand(selectedCmd, RunShellScriptCommand)

	// if len(sc.Coutput) != 0 {
	// 	// fmt.Println(sc.Coutput)

	// 	out, err := json.Marshal(sc)

	// 	if err != nil {
	// 		//w.WriteHeader(http.StatusBadRequest)
	// 		return

	// 	}
	// 	fmt.Println(string(out))

	// }

	// ret := UpdateCommandDuration(selectedCmd, sc.Duration)

	// if ret != true {
	// 	//w.WriteHeader(http.StatusBadRequest)
	// 	fmt.Fprintf(os.Stderr, "Error while updating command")
	// 	os.Exit(1)
	// }
}
