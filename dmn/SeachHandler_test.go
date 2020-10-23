package dmn

import "testing"

func TestSearchtHandler(t *testing.T) {

	err := InitHandlerTest()

	if err != nil {
		t.Errorf("Error initializing test %v", err)
	}

	// Create a command
	var cmd Command

	cmdStr := "ls -ltr"
	cmdDescription := "list files"

	cmd.Set(cmdStr, cmdDescription)

	// Manually populate our history file
	var requestHandler RequestHandler

	requestHandler.Set(TestSecret, TestHistory)
	ret := requestHandler.SaveCmd(cmd)

	if ret != true {
		t.Errorf("Unable to save command")
	}

	// Select the command. The hash is computed from the command string, so this is a well known constant.
	cmds, err := requestHandler.SearchCmd(cmdDescription)

	if err != nil {
		t.Errorf("Unable to select command")
	}

	if len(cmds) != 1 {
		t.Errorf("Command not found")
	}

	foundCmd := cmds[0]

	if foundCmd.CmdString != cmdStr || foundCmd.Description != cmdDescription {
		t.Errorf("Wrong command/description: %v: %v", foundCmd.CmdString, foundCmd.Description)
	}

}