package dmn

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// HandleSearch searches for a Command by its description. Only lowercase is used to evaulate
// whether a substring matches.
func (a *App) HandleSearch(w http.ResponseWriter, r *http.Request) {

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

	// Select the dmn.Command, otherwise, if the dmn.Command hash cannot be found, return error 400
	selectedCmds, cerr := a.SearchCmd(variables.Description)

	if cerr != nil {
		a.DmnLogFile.Log.Println("Unable to select Command")
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

// SearchCmd returns a Command by name
func (a *App) SearchCmd(description string) ([]Command, error) {

	a.DmnLogFile.Log.Println("Searching " + description)

	cmds, error := a.History.ReadCmdHistoryFile()

	ret := []Command{}

	if error != nil {
		return []Command{}, error
	}

	expectedDescription := strings.ToLower(description)

	for _, cmd := range cmds {

		// Use lower case for evaluation
		lowerDescription := strings.ToLower(cmd.Description)

		if strings.Contains(lowerDescription, expectedDescription) {
			ret = append(ret, cmd)
		}
	}

	return ret, nil
}
