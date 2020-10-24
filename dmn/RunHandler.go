package dmn

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// HandleRun runs a Command. If the command was run, we return status 200 regardless of
// whether it was run successfully or not, or whether there was an issue in writing the duration.
// This should be improved; however, the client still receives the ScedheduledCommand struct
// so any error messages will be there.
func (handler *RequestHandler) HandleRun(w http.ResponseWriter, r *http.Request) {

	log.Printf("Running command")

	// Get variables from the request
	vars := mux.Vars(r)
	var variables RequestVariable
	err := variables.GetVariablesFromRequestVars(vars)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the secret we passed in is valid, otherwise, return error 400
	if !handler.Secret.Valid(variables.Secret) {
		log.Println("Bad secret!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Select the dmn.Command, otherwise, if the dmn.Command hash cannot be found, return error 400
	selectedCmd, cerr := handler.SelectCmd(variables.CmdHash)

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

	var sc ScheduledCommand

	sc = handler.ScheduleCommand(selectedCmd)

	ret := handler.UpdateCommandDuration(selectedCmd, sc.Duration)

	if ret != true {
		log.Printf("Error in updating command duration\n")
	}

	log.Println("Command completed")

	w.WriteHeader(http.StatusOK)

	out, _ := json.Marshal(sc)

	io.WriteString(w, string(out))
}

// ScheduleCommand runs a Command based on a function passed in as the second parameter.
// This gives the ability to run Commands in multiple ways; for example, as a "mock" Command
// (RunMockCommand) or a shell script Command (RunShellScriptCommand).
func (handler *RequestHandler) ScheduleCommand(cmd Command) ScheduledCommand {

	var sc ScheduledCommand

	sc.CmdHash = cmd.CmdHash
	sc.CmdString = cmd.CmdString
	sc.Description = cmd.Description
	sc.Duration = -1

	log.Printf("Running command with hash: %v\n", sc.CmdHash)

	// Create a channel to hold exit status
	c := make(chan int)

	// Set the start time
	sc.StartTime = time.Now()

	// Run the Command in a goroutine
	go sc.RunShellScriptCommand(c)

	// Receive the exit status of the dmn.Command
	status := <-c

	now := time.Now()

	// Set end time after we receive from the channel
	sc.EndTime = now

	// Calculate the duration and store it
	sc.Duration = now.Sub(sc.StartTime)

	// The main reason why this code exists is to use the value received from the channel.
	if status != 0 {
		fmt.Fprintf(os.Stderr, "\nError: Command failed.\n")
	}

	return sc
}

// UpdateCommandDuration updates a Command with the same hash in the history file
func (handler *RequestHandler) UpdateCommandDuration(cmd Command, duration time.Duration) bool {

	log.Printf("Updating duration: %v\n", duration)

	// Check if the file does not exist. If not, then create it and add our first dmn.Command to it.
	f, err := os.Open(handler.History.Path)

	// Immediately close the file since we plan to write to it
	f.Close()

	// Check if the file doesn't exist and if so, then write it.
	if err != nil {

		// The array of dmn.Commands
		var cmds []Command

		// Set the duration
		cmd.Duration = duration

		cmds = append(cmds, cmd)

		mode := int(0644)

		updatedData, _ := json.MarshalIndent(cmds, "", "\t")

		error := ioutil.WriteFile(handler.History.Path, updatedData, os.FileMode(mode))

		return error == nil
	}

	// Update the dmn.Command in the history file

	// The array of dmn.Commands
	var cmds []Command

	// Read history file
	data, err := ioutil.ReadFile(handler.History.Path)

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

	error := ioutil.WriteFile(handler.History.Path, updatedData, os.FileMode(mode))

	return error == nil
}
