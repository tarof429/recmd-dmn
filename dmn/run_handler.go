package dmn

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
		handler.Log.Println("Bad secret!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Select the dmn.Command, otherwise, if the dmn.Command hash cannot be found, return error 400
	selectedCmd, cerr := handler.SelectCmd(variables.CmdHash)

	if cerr != nil {
		handler.Log.Println("Unable to select Command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if selectedCmd.CmdHash == "" {
		handler.Log.Println("Invalid hash")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handler.Log.Printf("Scheduling command %v: %v\n", selectedCmd.CmdHash, selectedCmd.Status)
	selectedCmd.Status = Scheduled
	handler.CommandScheduler.QueuedCommands = append(handler.CommandScheduler.QueuedCommands, selectedCmd)
	handler.CommandScheduler.CommandQueue <- selectedCmd

	handler.Log.Printf("Completed command %v: %v\n", selectedCmd.CmdHash, selectedCmd.Status)

	completedCommand := <-handler.CommandScheduler.CompletedQueue

	handler.Log.Printf("Command received from CompletedQueue: %v: %v\n", completedCommand.CmdHash, selectedCmd.Status)
	handler.UpdateCommandDuration(selectedCmd, completedCommand.Duration)

	for index, cmd := range handler.CommandScheduler.QueuedCommands {
		if cmd.CmdHash == selectedCmd.CmdHash {
			handler.Log.Printf("Updating command status%v: %v\n", selectedCmd.CmdHash, selectedCmd.Status)
			handler.CommandScheduler.QueuedCommands[index].Status = Completed

			handler.Log.Printf("Vacuuming command %v: %v\n", selectedCmd.CmdHash, selectedCmd.Status)
			handler.CommandScheduler.VacuumQueue <- selectedCmd
			break
		}
	}

	out, _ := json.Marshal(completedCommand)
	io.WriteString(w, string(out))
}

// UpdateCommandDuration updates a Command with the same hash in the history file
func (handler *RequestHandler) UpdateCommandDuration(cmd Command, duration time.Duration) bool {

	handler.Log.Printf("Updating %v: ran in %v\n", cmd.CmdHash, duration)

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