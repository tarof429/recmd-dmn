package dmn

import (
	"time"
)

// Scheduler manages commands and runs them
type Scheduler struct {
	CommandQueue   chan Command
	CompletedQueue chan ScheduledCommand
	QueuedCommands []Command
}

// CreateScheduler creates the channels
func (scheduler *Scheduler) CreateScheduler() {
	scheduler.CompletedQueue = make(chan ScheduledCommand)
	scheduler.CommandQueue = make(chan Command)
}

// Shutdown closes the channels
func (scheduler *Scheduler) Shutdown() {
	close(scheduler.CompletedQueue)
	close(scheduler.CommandQueue)
}

// RunSchedulerMock runs a mock schedule
func (scheduler *Scheduler) RunSchedulerMock() {

	for cmd := range scheduler.CommandQueue {
		//log.Printf("Scheduling command: %v\n", cmd.CmdHash)

		var sc ScheduledCommand
		sc.CmdHash = cmd.CmdHash

		time.Sleep(time.Second)

		//log.Printf("Command completed: %v\n", cmd.CmdHash)

		scheduler.CompletedQueue <- sc
	}
}

// RunScheduler reads off the CommandQueue and runs Commands
func (scheduler *Scheduler) RunScheduler() {
	for cmd := range scheduler.CommandQueue {
		//log.Printf("Scheduling command: %v\n", cmd.CmdHash)

		var sc ScheduledCommand

		sc.CmdHash = cmd.CmdHash
		sc.CmdString = cmd.CmdString
		sc.Description = cmd.Description
		sc.WorkingDirectory = cmd.WorkingDirectory
		sc.Status = Running
		sc.StartTime = time.Now()
		sc.RunShellScriptCommandWithExitStatus()
		sc.EndTime = time.Now()
		sc.Duration = sc.EndTime.Sub(sc.StartTime)
		sc.Status = Completed

		// if sc.ExitStatus != 0 {
		// 	log.Printf("Error: Command %v failed: %v\n", sc.CmdHash, sc.ExitStatus)
		// }

		scheduler.CompletedQueue <- sc
	}
}
