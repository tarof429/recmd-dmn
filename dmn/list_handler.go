package dmn

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

// HandleList lists Commands
func (a *App) HandleList(w http.ResponseWriter, r *http.Request) {

	a.RequestHandler.Log.Printf("Handling list")

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
		a.RequestHandler.Log.Printf("Bad secret! Expected %v but got t%v\n", a.RequestHandler.Secret.GetSecret(), variables.Secret)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cmds, err := a.ListCmd()

	if err != nil {
		a.RequestHandler.Log.Println("Unable to read history file")
	}

	w.WriteHeader(http.StatusOK)

	out, err := json.Marshal(cmds)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	io.WriteString(w, string(out))
}

// ListCmd lists Commands
func (a *App) ListCmd() ([]Command, error) {

	ret, err := a.RequestHandler.History.ReadCmdHistoryFile()

	return ret, err
}
