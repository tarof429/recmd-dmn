package dmn

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"time"
)

type CommandStatus string

const (
	Idle      CommandStatus = "Idle"
	Running   CommandStatus = "Running"
	Completed CommandStatus = "Completed"
	Scheduled CommandStatus = "Scheduled"
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
