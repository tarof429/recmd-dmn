package dmn

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)

func (handler *RequestHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {

	log.Println("Getting status")

	scs := handler.StatusCmd()

	// var scs []ScheduledCommand

	// for e := list.Front(); e != nil; e = e.Next() {

	// 	cmd := ScheduledCommand(e.Value)

	// 	scs = append(scs, cmd)
	// 	//fmt.Println(e.Value)
	// }

	w.WriteHeader(http.StatusOK)

	out, _ := json.Marshal(scs)

	io.WriteString(w, string(out))
}

func (handler *RequestHandler) StatusCmd() []ScheduledCommand {

	//var ret []ScheduledCommand

	// var sc ScheduledCommand
	// sc.CmdString = "foo"
	// sc.CmdHash = "642618de1bfe68c92e089a092eebb7"
	// sc.Status = "Running"

	// ret = append(ret, sc)

	queue := handler.CommandScheduler.QueuedCommands

	log.Println("Total queued: " + strconv.Itoa(len(queue)))

	// for e := queue.Front(); e != nil; e = e.Next() {
	// 	ret = append(ret, e)
	// }

	return queue

}
