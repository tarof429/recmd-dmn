package dmn

import (
	"crypto/sha1"
	"fmt"
	"testing"
	"time"
)

func TestScheduler(t *testing.T) {

	formattedHash := func(cmdString string) string {
		h := sha1.New()
		h.Write([]byte(cmdString))
		return fmt.Sprintf("%.15x", h.Sum(nil))
	}

	var scheduler Scheduler
	scheduler.CreateScheduler()

	var cmd1 Command
	cmd1.CmdString = "#This is a command"
	cmd1.CmdHash = formattedHash(cmd1.CmdString)

	var cmd2 Command
	cmd2.CmdString = "#This is another command"
	cmd2.CmdHash = formattedHash(cmd2.CmdString)

	go scheduler.Schedule(cmd1)
	go scheduler.Schedule(cmd2)

	go func() {
		time.Sleep(time.Second * 3)
		scheduler.Shutdown()
	}()

	scheduler.RunSchedulerMock()

}
