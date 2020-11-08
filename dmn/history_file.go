package dmn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	// The Command history file
	recmdHistoryFile = "recmd_history.json"
)

// HistoryFile represents the file containing the history
type HistoryFile struct {
	Path string
}

// Set sets the path to the history file
func (h *HistoryFile) Set(path string) {
	h.Path = filepath.Join(path, recmdHistoryFile)
}

// Get returns the path to the history file
func (h *HistoryFile) Get() string {
	return h.Path
}

// Remove removes the history file
func (h *HistoryFile) Remove() {
	os.Remove(h.Path)
}

// Create creates the history file
func (h *HistoryFile) Create() {
	os.Create((h.Path))
}

// ReadCmdHistoryFile reads historyFile and generates a list of Command structs
func (h *HistoryFile) ReadCmdHistoryFile() ([]Command, error) {

	var (
		historyData []byte    // Data representing our history file
		cmds        []Command // List of dmn.Commands produced after unmarshalling historyData
		err         error     // Any errors we might encounter
	)

	// Read the history file into historyData
	historyData, err = ioutil.ReadFile(h.Path)

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

// OverwriteCmdHistoryFile overwrites the history file with []dmn.Command passed in as a parameter
func (h *HistoryFile) OverwriteCmdHistoryFile(cmds []Command) bool {

	mode := int(0644)

	updatedData, _ := json.MarshalIndent(cmds, "", "\t")

	error := ioutil.WriteFile(h.Path, updatedData, os.FileMode(mode))

	return error == nil
}

// WriteHistoryToFile creates an empty history file
func (h *HistoryFile) WriteHistoryToFile() error {

	// Check if the file does not exist. If not, then create it and add our first dmn.Command to it.
	f, err := os.Open(h.Path)

	// Immediately close the file since we plan to write to it
	defer f.Close()

	// Check if the file doesn't exist and if so, then write it.
	if err != nil {

		mode := int(0644)

		error := ioutil.WriteFile(h.Path, []byte(nil), os.FileMode(mode))

		if err != nil {
			return error
		}
	}
	return nil
}
