package dmn

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func (handler *RequestHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {

	handler.Log.Println("Handling status")

	// Get variables from the request
	vars := mux.Vars(r)
	var variables RequestVariable
	err := variables.GetVariablesFromRequestVars(vars)

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

	scs := handler.StatusCmd()

	w.WriteHeader(http.StatusOK)

	out, _ := json.Marshal(scs)

	io.WriteString(w, string(out))
}

func (handler *RequestHandler) StatusCmd() []Command {

	cmds := handler.CommandScheduler.QueuedCommands

	//handler.Log.Println("Total queued: " + strconv.Itoa(len(cmds)))

	return cmds

}
