package dmn

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// HandleShow shows the actual command that will be run.
// If the script is inline, just return that string.
// However if it points to a script, get the contents of that
// script and return it as a string.
func (a *App) HandleShow(w http.ResponseWriter, r *http.Request) {

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
	if !a.Secret.Valid(variables.Secret) {
		a.DmnLogFile.Log.Println("Bad secret!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Select the Command, otherwise, if the Command hash cannot be found, return error 400
	selectedCmd, cerr := a.SelectCmd(variables.CmdHash)

	if cerr != nil {
		a.DmnLogFile.Log.Println("Unable to select Command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if selectedCmd.CmdHash == "" {
		a.DmnLogFile.Log.Println("Invalid hash")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cmdString := strings.Split(selectedCmd.CmdString, " ")

	wd := selectedCmd.WorkingDirectory

	var ret string

	a.DmnLogFile.Log.Printf("Command string is: %v\n", cmdString)

	if len(cmdString) == 2 {

		pathToScript := filepath.Join(wd, cmdString[1])
		a.DmnLogFile.Log.Println("Path to script: " + pathToScript)

		if _, err := os.Stat(pathToScript); err == nil {
			fileData, err := ioutil.ReadFile(pathToScript)

			if err != nil {
				a.DmnLogFile.Log.Printf("An error occurred while reading historyfile: %v\n", err.Error())
			} else {
				a.DmnLogFile.Log.Printf("Returning contents of %v\n", pathToScript)
				ret = string(fileData)
			}

		}
	} else {
		ret = selectedCmd.CmdString
	}

	w.WriteHeader(http.StatusOK)

	out, err := json.Marshal(ret)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	io.WriteString(w, string(out))
}
