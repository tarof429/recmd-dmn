package dmn

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *App) HandleStatus(w http.ResponseWriter, r *http.Request) {

	a.DmnLogFile.Log.Println("Handling status")

	// Get variables from the request
	vars := mux.Vars(r)
	var variables RequestVariable
	err := variables.GetVariablesFromRequestVars(vars)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the secret we passed in is valid, otherwise, return error 400
	if !a.Secret.Valid(variables.Secret) {
		a.DmnLogFile.Log.Printf("Bad secret! Expected %v but got t%v\n", a.Secret.GetSecret(), variables.Secret)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	scs := a.StatusCmd()

	w.WriteHeader(http.StatusOK)

	out, _ := json.Marshal(scs)

	io.WriteString(w, string(out))
}

// StatusCmd returns a list of queued commands
func (a *App) StatusCmd() []Command {

	cmds := a.RequestHandler.CommandScheduler.QueuedCommands

	a.DmnLogFile.Log.Println("Total queued: " + strconv.Itoa(len(cmds)))

	return cmds

}
