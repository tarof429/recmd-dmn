package dmn

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func (a *App) HandleStatus(w http.ResponseWriter, r *http.Request) {

	a.RequestHandler.Log.Println("Handling status")

	// Get variables from the request
	vars := mux.Vars(r)
	var variables RequestVariable
	err := variables.GetVariablesFromRequestVars(vars)

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

	scs := a.StatusCmd()

	w.WriteHeader(http.StatusOK)

	out, _ := json.Marshal(scs)

	io.WriteString(w, string(out))
}

func (a *App) StatusCmd() []Command {

	cmds := a.RequestHandler.CommandScheduler.QueuedCommands

	//handler.Log.Println("Total queued: " + strconv.Itoa(len(cmds)))

	return cmds

}
