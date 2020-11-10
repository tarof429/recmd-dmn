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

	var a App

	a.CreateScheduler()
	a.DmnLogFile.Set("testdata")
	a.DmnLogFile.Create()

	var cmd1 Command
	cmd1.CmdString = "# This is a command"
	cmd1.CmdHash = formattedHash(cmd1.CmdString)
	cmd1.Status = Scheduled

	var cmd2 Command
	cmd2.CmdString = "# This is another command"
	cmd2.CmdHash = formattedHash(cmd2.CmdString)
	cmd2.Status = Scheduled

	// Create a goroutine to continuously read from the CompletedQueue channel
	go func() {
		for sc := range a.CommandScheduler.CompletedQueue {
			//fmt.Printf("Command received from CompletedQueue: %v %v %v\n", sc.CmdHash, sc.Description, sc.Status)

			for index, cmd := range a.CommandScheduler.QueuedCommands {
				if cmd.CmdHash == sc.CmdHash {
					fmt.Printf("Updating status of %v: %v\n", cmd.CmdHash, Completed)
					a.CommandScheduler.QueuedCommands[index].Status = Completed
					break
				}
			}
		}
	}()

	// Schedule the Commands. This will take Commands off of the CommandQueue, run them, and put the ScheduledCommands onto the CompletedQueue
	go a.RunSchedulerMock()

	// Add commands to the array of queued commands. This is used to track the state.
	a.CommandScheduler.QueuedCommands = append(a.CommandScheduler.QueuedCommands, cmd1)
	a.CommandScheduler.QueuedCommands = append(a.CommandScheduler.QueuedCommands, cmd2)

	// Now feed the CommandQueue
	a.CommandScheduler.CommandQueue <- cmd1
	a.CommandScheduler.CommandQueue <- cmd2

	time.Sleep(time.Second * 3)
}
