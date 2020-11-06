package dmn

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// HandleDelete deletes a Command
func (a *App) HandleDelete(w http.ResponseWriter, r *http.Request) {

	a.RequestHandler.Log.Println("Handling delete")

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
	if !a.RequestHandler.Secret.Valid(variables.Secret) {
		a.RequestHandler.Log.Println("Bad secret!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Select the dmn.Command, otherwise, if the dmn.Command hash cannot be found, return error 400
	selectedCmd, err := a.DeleteCmd(variables.CmdHash)

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

// DeleteCmd deletes a dmn.Command. It's best to pass in the dmn.CommandHash
// because dmn.Commands may look similar.
func (a *App) DeleteCmd(value string) ([]Command, error) {

	a.RequestHandler.Log.Println("Deleting " + value)

	ret := []Command{}

	cmds, error := a.RequestHandler.History.ReadCmdHistoryFile()

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
		a.RequestHandler.History.OverwriteCmdHistoryFile(cmds)
	}

	return ret, nil
}
