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

	defer scheduler.Shutdown() // Shutdown the scheduler to close the channels

	var cmd1 Command
	cmd1.CmdString = "#This is a command"
	cmd1.CmdHash = formattedHash(cmd1.CmdString)

	var cmd2 Command
	cmd2.CmdString = "#This is another command"
	cmd2.CmdHash = formattedHash(cmd2.CmdString)

	// Create a goroutine to continuously read from the CompletedQueue channel
	go func() {
		for sc := range scheduler.CompletedQueue {
			fmt.Printf("Command received from CompletedQueue: %v\n", sc.CmdHash)
		}
	}()

	// Schedule the Commands. This will take Commands off of the CommandQueue, run them, and put the ScheduledCommands onto the CompletedQueue
	go scheduler.RunSchedulerMock()

	// Now feed the CommandQueue
	scheduler.CommandQueue <- cmd1
	scheduler.CommandQueue <- cmd2

	time.Sleep(time.Second * 3)
}
