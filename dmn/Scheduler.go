package dmn

import (
	"fmt"
	"os"
	"time"
)

// Scheduler manages commands and runs them
type Scheduler struct {
	Queue          chan ScheduledCommand
	CompletedQueue chan ScheduledCommand
}

// CreateScheduler creates the channels
func (scheduler *Scheduler) CreateScheduler() {
	scheduler.Queue = make(chan ScheduledCommand)
	scheduler.CompletedQueue = make(chan ScheduledCommand)
}

// Schedule schdules a command and puts it into the queue
func (scheduler *Scheduler) Schedule(cmd Command) {

	var sc ScheduledCommand

	sc.CmdHash = cmd.CmdHash
	sc.CmdString = cmd.CmdString
	sc.Description = cmd.Description
	sc.Duration = -1
	sc.Status = Queued

	scheduler.Queue <- sc
}

// Shutdown closes the channels
func (scheduler *Scheduler) Shutdown() {
	close(scheduler.Queue)
	close(scheduler.CompletedQueue)
}

// RunSchedulerMock runs a mock schedule
func (scheduler *Scheduler) RunSchedulerMock() {

	for sc := range scheduler.Queue {
		fmt.Println("Received new command: " + sc.CmdHash)
		time.Sleep(time.Second)
		fmt.Println("Command completed")
	}
}

// RunScheduler runs the schedule
func (scheduler *Scheduler) RunScheduler() {

	for sc := range scheduler.Queue {
		sc.Status = Running
		sc.StartTime = time.Now()
		sc.RunShellScriptCommandWithExitStatus()
		sc.EndTime = time.Now()
		sc.Duration = sc.EndTime.Sub(sc.StartTime)
		sc.Status = Completed

		if sc.ExitStatus != 0 {
			fmt.Fprintf(os.Stderr, "\nError: Command failed.\n")
		}
		scheduler.CompletedQueue <- sc
	}
}
