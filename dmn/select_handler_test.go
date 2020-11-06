package dmn

import (
	"fmt"
	"testing"
)

func TestSelectHandler(t *testing.T) {

	var app App

	err := app.InitHandlerTest()

	if err != nil {
		t.Errorf("Error initializing test %v", err)
	}

	// Create a command
	var cmd Command

	cmdStr := "ls -ltr"
	cmdDescription := "list files"
	workingDirectory := "."

	cmd.Set(cmdStr, cmdDescription, workingDirectory)
	expectedCommandHash := cmd.CmdHash

	fmt.Println("Looking for " + expectedCommandHash)

	// // Manually populate our history file
	// var requestHandler RequestHandler

	// requestHandler.Set(TestSecret, TestHistory)
	// requestHandler.Log = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	ret := app.RequestHandler.SaveCmd(cmd)

	if ret != true {
		t.Errorf("Unable to save command")
	}

	// Select the command. The hash is computed from the command string, so this is a well known constant.
	cmd, err = app.RequestHandler.SelectCmd(expectedCommandHash)

	if err != nil {
		t.Errorf("Unable to select command")
	}

	if cmd.CmdString != cmdStr || cmd.Description != cmdDescription {
		t.Errorf("Wrong command/description: %v: %v", cmd.CmdString, cmd.Description)
	}

}
