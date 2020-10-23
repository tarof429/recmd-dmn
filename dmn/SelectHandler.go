package dmn

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// HandleSelect selects a Command
func (handler *RequestHandler) HandleSelect(w http.ResponseWriter, r *http.Request) {

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

	// Select the Command, otherwise, if the Command hash cannot be found, return error 400
	selectedCmd, cerr := handler.SelectCmd(variables.CmdHash)

	if cerr != nil {
		log.Println("Unable to select Command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if selectedCmd.CmdHash == "" {
		log.Println("Invalid hash")
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
func (handler *RequestHandler) SelectCmd(value string) (Command, error) {

	log.Println("Selecting " + value)

	cmds, error := handler.History.ReadCmdHistoryFile()

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
