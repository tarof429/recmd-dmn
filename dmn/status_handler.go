package dmn

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

// HandleStatus handles the status command
func (a *App) HandleStatus(w http.ResponseWriter, r *http.Request) {

	// Get variables from the request
	vars := mux.Vars(r)
	var variables RequestVariable
	err := variables.GetVariablesFromRequestVars(vars)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Check if the secret we passed in is valid, otherwise, return error 400
	if !a.Secret.Valid(variables.Secret) {
		a.DmnLogFile.Log.Println("Bad secret!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Select the Command, otherwise, if the Command hash cannot be found, return error 400
	status, cerr := a.StatusCmd()

	if cerr != nil {
		a.DmnLogFile.Log.Println("Unable to get status")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	out, err := json.Marshal(status)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	io.WriteString(w, string(out))
}

// StatusCmd indicates if the app is up or down. This always returns true.
func (a *App) StatusCmd() (bool, error) {

	a.DmnLogFile.Log.Println("Getting status")

	return true, nil
}
