package dmn

import (
	"fmt"
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
func (a *App) CreateScheduler() {
	a.CommandScheduler.CompletedQueue = make(chan ScheduledCommand)
	a.CommandScheduler.CommandQueue = make(chan Command)
	a.CommandScheduler.VacuumQueue = make(chan Command)
}

// QueuedCommandsCleanup removes completed commands from the array.
// Once it receives a command from the channel, it starts a goroutine
// to make it a non-blocking operation.
func (a *App) QueuedCommandsCleanup() {
	for selectedCmd := range a.CommandScheduler.VacuumQueue {
		go func() {
			time.Sleep(time.Second * 3)
			a.DmnLogFile.Log.Printf("Vacuuming %v\n", selectedCmd.CmdHash)
			for foundIndex, cmd := range a.CommandScheduler.QueuedCommands {
				if cmd.CmdHash == selectedCmd.CmdHash {
					//log.Printf("Vacuuming command: %v\n", cmd.CmdHash)
					a.CommandScheduler.QueuedCommands = append(a.CommandScheduler.QueuedCommands[:foundIndex], a.CommandScheduler.QueuedCommands[foundIndex+1:]...)
					break
				}
			}
		}()
	}
	a.DmnLogFile.Log.Printf("Total queued: %v\n", len(a.CommandScheduler.QueuedCommands))
}

func (a *App) updateStatusForQueuedCommand(selectedCmd Command, status CommandStatus) {
	for foundIndex, cmd := range a.CommandScheduler.QueuedCommands {
		if cmd.CmdHash == selectedCmd.CmdHash {
			a.DmnLogFile.Log.Printf("Updating status for %v: %v to %v\n", cmd.CmdHash, a.CommandScheduler.QueuedCommands[foundIndex].Status, status)
			a.CommandScheduler.QueuedCommands[foundIndex].Status = status
			break
		}
	}
}

// RunSchedulerMock runs a mock schedule
func (a *App) RunSchedulerMock(expectedStatus CommandStatus) {

	fmt.Println("In RunSchedulerMock")

	for cmd := range a.CommandScheduler.CommandQueue {
		fmt.Printf("Scheduling command: %v\n", cmd.CmdHash)

		var sc ScheduledCommand
		sc.CmdHash = cmd.CmdHash
		sc.Description = cmd.Description
		sc.WorkingDirectory = cmd.WorkingDirectory
		sc.Status = Running
		a.updateStatusForQueuedCommand(cmd, Running)

		sc.StartTime = time.Now()
		sc.RunShellScriptCommandWithExpectedStatus(expectedStatus)
		fmt.Printf("Completed running command")
		//time.Sleep(time.Second * 1) // Simulate command execution

		// Simulate working directory not present
		// sc.WorkingDirectory

		sc.EndTime = time.Now()
		sc.Duration = sc.EndTime.Sub(sc.StartTime)

		//sc.Status = Completed
		a.updateStatusForQueuedCommand(cmd, sc.Status)

		a.DmnLogFile.Log.Printf("Command completed: %v\n", cmd.CmdHash)
		fmt.Printf("Command completed: %v\n", cmd.CmdHash)

		a.CommandScheduler.CompletedQueue <- sc
	}
}

// RunScheduler reads off the CommandQueue and runs Commands
func (a *App) RunScheduler() {
	for cmd := range a.CommandScheduler.CommandQueue {
		//log.Printf("Scheduling command: %v\n", cmd.CmdHash)

		var sc ScheduledCommand

		sc.CmdHash = cmd.CmdHash
		sc.CmdString = cmd.CmdString
		sc.Description = cmd.Description
		sc.WorkingDirectory = cmd.WorkingDirectory
		sc.Status = Running
		a.updateStatusForQueuedCommand(cmd, Running)

		sc.StartTime = time.Now()
		sc.RunShellScriptCommandWithExitStatus()
		sc.EndTime = time.Now()
		sc.Duration = sc.EndTime.Sub(sc.StartTime)
		//sc.Status = Completed
		a.updateStatusForQueuedCommand(cmd, sc.Status)

		if sc.Status != Completed {
			a.DmnLogFile.Log.Printf("Error: Command %v failed with exit status %v: %v\n", sc.CmdHash, sc.ExitStatus, sc.Coutput)
		}

		a.CommandScheduler.CompletedQueue <- sc
	}
}
