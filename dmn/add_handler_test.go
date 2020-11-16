package dmn

import (
	"testing"
)

func TestAddHandler(t *testing.T) {

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

}

func TestSaveCmdWithBadWorkingDirectory(t *testing.T) {

	var app App

	err := app.InitalizeTest()

	if err != nil {
		t.Errorf("Error initializing test %v", err)
	}

	var cmd Command

	cmd.Set("another-command", "another-command", "another-command")

	ret := app.SaveCmd(cmd)

	if ret != false {
		t.Errorf("Saved command with bad working diectory")
	}
}
