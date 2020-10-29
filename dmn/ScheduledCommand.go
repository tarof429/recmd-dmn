package dmn

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

type ScheduledStatus string

const (
	Running   ScheduledStatus = "Running"
	Queued    ScheduledStatus = "Queued"
	Completed ScheduledStatus = "Completed"
)

// ScheduledCommand represents a Command that is scheduled to run
type ScheduledCommand struct {
	Command
	Coutput    string    `json:"coutput"`
	ExitStatus int       `json:"exitStatus"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
	Status     ScheduledStatus
}

// RunShellScriptCommandWithExitStatus runs a Command written to a temporary file
func (sc *ScheduledCommand) RunShellScriptCommandWithExitStatus() int {

	// log.Printf("In RunShellScriptCommand")
	// log.Printf("Will run: %v\n", sc.CmdString)

	tempFile, err := ioutil.TempFile(os.TempDir(), "recmd-")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to create temp file: %d\n", err)
	}

	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("#!/bin/sh\n\n" + sc.CmdString)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Errror: unable to write script to temp file: : %s\n", err)
	}

	cmd := exec.Command("sh", tempFile.Name())

	// We may want to make this configurable in the future.
	// For now, all dmn.Commands will be run from the user's home directory
	cmd.Dir, err = os.UserHomeDir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to obtain home directory: %s\n", err)
	}

	//out, err := cmd.Output()

	combinedOutput, combinedOutputErr := cmd.CombinedOutput()

	// fmt.Fprintf(os.Stdout, "\nError: %s error 2: %v\n", string(combinedOutput), err2)

	if combinedOutputErr != nil {
		sc.ExitStatus = -1
	}

	sc.Coutput = string(combinedOutput)

	return sc.ExitStatus
}
