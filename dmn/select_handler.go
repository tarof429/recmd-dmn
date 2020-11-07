package dmn

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// HandleSelect selects a Command by its hash
func (a *App) HandleSelect(w http.ResponseWriter, r *http.Request) {

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

	w.WriteHeader(http.StatusOK)

	out, err := json.Marshal(selectedCmd)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	io.WriteString(w, string(out))
}

// SelectCmd returns a Command
func (a *App) SelectCmd(value string) (Command, error) {

	a.DmnLogFile.Log.Println("Selecting " + value)

	cmds, error := a.History.ReadCmdHistoryFile()

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
