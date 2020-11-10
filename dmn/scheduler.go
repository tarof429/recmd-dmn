package dmn

import (
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

// QueuedCommandsCleanup removes completed commands from the array
func (a *App) QueuedCommandsCleanup() {
	for selectedCmd := range a.CommandScheduler.VacuumQueue {
		time.Sleep(time.Second * 10)
		a.DmnLogFile.Log.Printf("Vacuuming %v\n", selectedCmd.CmdHash)
		for foundIndex, cmd := range a.CommandScheduler.QueuedCommands {
			if cmd.CmdHash == selectedCmd.CmdHash {
				//log.Printf("Vacuuming command: %v\n", cmd.CmdHash)
				a.CommandScheduler.QueuedCommands = append(a.CommandScheduler.QueuedCommands[:foundIndex], a.CommandScheduler.QueuedCommands[foundIndex+1:]...)
				break
			}
		}
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
func (a *App) RunSchedulerMock() {

	for cmd := range a.CommandScheduler.CommandQueue {
		a.DmnLogFile.Log.Printf("Scheduling command: %v\n", cmd.CmdHash)

		var sc ScheduledCommand
		sc.CmdHash = cmd.CmdHash
		sc.Description = cmd.Description
		sc.WorkingDirectory = cmd.WorkingDirectory
		sc.Status = Running
		a.updateStatusForQueuedCommand(cmd, Running)

		sc.StartTime = time.Now()
		time.Sleep(time.Second * 1) // Simulate command execution
		sc.EndTime = time.Now()
		sc.Duration = sc.EndTime.Sub(sc.StartTime)
		sc.Status = Completed
		a.updateStatusForQueuedCommand(cmd, Completed)

		a.DmnLogFile.Log.Printf("Command completed: %v\n", cmd.CmdHash)

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
		sc.Status = Completed
		a.updateStatusForQueuedCommand(cmd, Completed)

		if sc.ExitStatus != 0 {
			a.DmnLogFile.Log.Printf("Error: Command %v failed with exit status %v: %v\n", sc.CmdHash, sc.ExitStatus, sc.Coutput)
		}

		a.CommandScheduler.CompletedQueue <- sc
	}
}
