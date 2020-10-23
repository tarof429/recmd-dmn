package dmn

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// HandleSearch searches for a Command by its description. Only lowercase is used to evaulate
// whether a substring matches.
func (handler *RequestHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {

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
	selectedCmds, cerr := handler.SearchCmd(variables.Description)

	if cerr != nil {
		log.Println("Unable to select Command")
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

// SearchCmd returns a dmn.Command by name
func (handler *RequestHandler) SearchCmd(value string) ([]Command, error) {

	log.Println("Searching " + value)

	cmds, error := handler.History.ReadCmdHistoryFile()

	ret := []Command{}

	if error != nil {
		return []Command{}, error
	}

	for _, cmd := range cmds {

		// Use lower case for evaluation
		description := strings.ToLower(cmd.Description)

		if strings.Contains(description, value) {
			ret = append(ret, cmd)
		}
	}

	return ret, nil
}
