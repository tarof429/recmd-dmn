package dmn

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

// HandleList lists Commands
func (handler *RequestHandler) HandleList(w http.ResponseWriter, r *http.Request) {

	handler.Log.Printf("Handling list")

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
		handler.Log.Printf("Bad secret! Expected %v but got t%v\n", handler.Secret.GetSecret(), variables.Secret)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cmds, err := handler.ListCmd()

	if err != nil {
		handler.Log.Println("Unable to read history file")
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
func (handler *RequestHandler) ListCmd() ([]Command, error) {

	ret, err := handler.History.ReadCmdHistoryFile()

	return ret, err
}
