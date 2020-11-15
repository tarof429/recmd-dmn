package dmn

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"time"
)

// CommandStatus indicates the status of the command
type CommandStatus string

const (
	// Idle means that the command is not running
	Idle CommandStatus = "Idle"

	// Running means that the command is running
	Running CommandStatus = "Running"

	// Completed means that the command is done
	Completed CommandStatus = "Completed"

	// Scheduled means that the command will run
	Scheduled CommandStatus = "Scheduled"

	// Failed means that the command failed
	Failed CommandStatus = "Failed"
)

// Command represents a command and optionally a description to document what the command does
type Command struct {
	CmdHash          string        `json:"commandHash"`
	CmdString        string        `json:"commandString"`
	Description      string        `json:"description"`
	Duration         time.Duration `json:"duration"`
	WorkingDirectory string        `json:"workingDirectory"`
	Status           CommandStatus `json:"status"`
}

// Set sets the fields of a new Command
func (cmd *Command) Set(cmdString string, cmdComment string, workingDirectory string) {

	formattedHash := func() string {
		h := sha1.New()
		h.Write([]byte(cmdString))
		return fmt.Sprintf("%.15x", h.Sum(nil))
	}()

	cmd.CmdHash = formattedHash
	cmd.CmdString = strings.Trim(cmdString, "")
	cmd.Description = strings.Trim(cmdComment, "")
	cmd.WorkingDirectory = strings.Trim(workingDirectory, "")
	cmd.Duration = -1
	cmd.Status = Idle
}
