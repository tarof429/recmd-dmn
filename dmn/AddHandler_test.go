package dmn

import (
	"testing"
)

func TestAddHandler(t *testing.T) {

	InitHandlerTest()

	// Create a command
	var cmd Command
	cmd.Set("ls", "list files")

	// Manually populate our history file
	var addHandler AddHandler

	addHandler.Set(TestSecret, TestHistory)
	ret := addHandler.SaveCmd(cmd)

	if ret != true {
		t.Errorf("Unable to save command")
	}

}
