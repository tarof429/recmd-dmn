package dmn

import (
	"testing"
)

func TestListHandler(t *testing.T) {

	var app App

	err := app.InitalizeTest()

	if err != nil {
		t.Errorf("Error initializing test %v", err)
	}

	// Create a command
	var cmd Command
	cmd.Set("ls", "list files", ".")

	ret := app.SaveCmd(cmd)

	if ret != true {
		t.Errorf("Unable to save command")
	}

	cmds, err := app.ListCmd()

	if err != nil {
		t.Errorf("Error when listing commands")
	}

	if len(cmds) != 1 {
		t.Errorf("No commands to list")
	}
}
