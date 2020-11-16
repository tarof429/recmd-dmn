package dmn

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

// ScheduledCommand represents a Command that is scheduled to run
type ScheduledCommand struct {
	Command
	Coutput    string    `json:"coutput"`
	ExitStatus int       `json:"exitStatus"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
}

func getCurrentWorkingDirectory() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Unable to obtain current working directory")
	}
	return cwd
}
func (sc *ScheduledCommand) RunShellScriptCommandWithExpectedStatus(expectedStatus CommandStatus) {
	fmt.Println("In RunShellScriptCommandWithExpectedStatus")
	time.Sleep(time.Second)
	sc.Status = expectedStatus
	fmt.Println("Completed RunShellScriptCommandWithExpectedStatus")
}

// RunShellScriptCommandWithExitStatus runs a Command written to a temporary file
func (sc *ScheduledCommand) RunShellScriptCommandWithExitStatus() int {

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
	cmd.Dir = sc.WorkingDirectory

	// Set a default working directory if it's not set
	if sc.WorkingDirectory == "" {
		sc.WorkingDirectory = getCurrentWorkingDirectory()
	}
	cmd.Dir = sc.WorkingDirectory

	combinedOutput, combinedOutputErr := cmd.CombinedOutput()

	// fmt.Fprintf(os.Stdout, "\nError: %s error 2: %v\n", string(combinedOutput), err2)

	if combinedOutputErr != nil {
		//sc.ExitStatus = -1
		sc.Status = Failed
	} else {
		sc.Status = Completed
	}

	sc.Coutput = string(combinedOutput)

	return sc.ExitStatus
}
