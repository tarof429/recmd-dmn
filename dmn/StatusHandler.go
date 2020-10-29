package dmn

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

func (handler *RequestHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {

	handler.Log.Println("Getting status")

	scs := handler.StatusCmd()

	w.WriteHeader(http.StatusOK)

	out, _ := json.Marshal(scs)

	io.WriteString(w, string(out))
}

func (handler *RequestHandler) StatusCmd() []ScheduledCommand {

	queue := handler.CommandScheduler.QueuedCommands

	handler.Log.Println("Total queued: " + strconv.Itoa(len(queue)))

	return queue

}
