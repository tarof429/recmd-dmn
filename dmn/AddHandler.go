package dmn

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type AddHandler struct {
	Secret  Secret
	History HistoryFile
}

// Set sets some variables
func (handler *AddHandler) Set(secret Secret, history HistoryFile) {
	handler.Secret = secret
	handler.History = history
}

// Handle adds a Command
func (handler *AddHandler) Handle(w http.ResponseWriter, r *http.Request) {

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
	testCmd := new(Command)

	testCmd.Set(variables.Command, variables.Description)

	if handler.SaveCmd(*testCmd) != true {
		w.WriteHeader(http.StatusBadRequest)
		out, _ := json.Marshal("false")
		io.WriteString(w, string(out))
		return
	}

	w.WriteHeader(http.StatusOK)
	out, _ := json.Marshal("true")
	io.WriteString(w, string(out))
}

// SaveCmd writes a dmn.Command to the history file
func (handler *AddHandler) SaveCmd(cmd Command) bool {

	// Check if the file does not exist. If not, then create it and add our first dmn.Command to it.
	f, err := os.Open(handler.History.Path)

	// Immediately close the file since we plan to write to it
	f.Close()

	// Check if the file doesn't exist and if so, then write it.
	if err != nil {
		// The array of dmn.Commands
		var cmds []Command

		cmds = append(cmds, cmd)

		mode := int(0644)

		updatedData, _ := json.MarshalIndent(cmds, "", "\t")

		error := ioutil.WriteFile(handler.History.Path, updatedData, os.FileMode(mode))

		return error == nil
	}

	// Update the Command in the history file

	// The array of dmn.Commands
	var cmds []Command

	// Read history file
	data, err := ioutil.ReadFile(handler.History.Path)

	// An error occured while reading historyFile.
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return false
	}

	// No data in file, write our Command to it
	if len(data) == 0 {
		cmds = append(cmds, cmd)
		updatedData, _ := json.MarshalIndent(cmds, "", "\t")
		mode := int(0644)
		error := ioutil.WriteFile(handler.History.Path, updatedData, os.FileMode(mode))
		return error == nil
	}
	if err := json.Unmarshal(data, &cmds); err != nil {
		fmt.Fprintf(os.Stderr, "JSON unmarshalling failed: %s\n", err)
		return false
	}

	// Check if the dmn.Command hash alaready exists, and prevent the user from adding the same Command
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

	error := ioutil.WriteFile(handler.History.Path, updatedData, os.FileMode(mode))

	return error == nil

}
