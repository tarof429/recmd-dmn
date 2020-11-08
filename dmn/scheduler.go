package dmn

import (
	"log"
	"time"
)

// Scheduler manages commands and runs them
type Scheduler struct {
	CommandQueue   chan Command
	CompletedQueue chan ScheduledCommand
	VacuumQueue    chan Command
	QueuedCommands []Command
}

// CreateScheduler creates the channels
func (scheduler *Scheduler) CreateScheduler() {
	scheduler.CompletedQueue = make(chan ScheduledCommand)
	scheduler.CommandQueue = make(chan Command)
	scheduler.VacuumQueue = make(chan Command)
}

// QueuedCommandsCleanup removes completed commands from the array
func (scheduler *Scheduler) QueuedCommandsCleanup() {
	for selectedCmd := range scheduler.VacuumQueue {
		for foundIndex, cmd := range scheduler.QueuedCommands {
			if cmd.CmdHash == selectedCmd.CmdHash {
				//log.Printf("Vacuuming command: %v\n", cmd.CmdHash)
				scheduler.QueuedCommands = append(scheduler.QueuedCommands[:foundIndex], scheduler.QueuedCommands[foundIndex+1:]...)
				break
			}
		}
		time.Sleep(time.Second * 10)
	}
}

// RunSchedulerMock runs a mock schedule
func (scheduler *Scheduler) RunSchedulerMock() {

	for cmd := range scheduler.CommandQueue {
		log.Printf("Scheduling command: %v\n", cmd.CmdHash)

		var sc ScheduledCommand
		sc.CmdHash = cmd.CmdHash
		sc.Description = cmd.Description
		sc.WorkingDirectory = cmd.WorkingDirectory
		sc.Status = Running
		sc.StartTime = time.Now()
		time.Sleep(time.Second * 1) // Simulate command execution
		sc.EndTime = time.Now()
		sc.Duration = sc.EndTime.Sub(sc.StartTime)
		sc.Status = Completed

		log.Printf("Command completed: %v\n", cmd.CmdHash)

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
