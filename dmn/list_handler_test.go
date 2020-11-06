package dmn

import (
	"testing"
)

func TestListHandler(t *testing.T) {

	var app App

	err := app.InitHandlerTest()

	if err != nil {
		t.Errorf("Error initializing test %v", err)
	}

	// Create a command
	var cmd Command
	cmd.Set("ls", "list files", ".")

	// // Manually populate our history file
	// var requestHandler RequestHandler

	// requestHandler.Set(TestSecret, TestHistory)
	// requestHandler.Log = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	ret := app.RequestHandler.SaveCmd(cmd)

	if ret != true {
		t.Errorf("Unable to save command")
	}

	cmds, err := app.RequestHandler.ListCmd()

	if err != nil {
		t.Errorf("Error when listing commands")
	}

	if len(cmds) != 1 {
		t.Errorf("No commands to list")
	}
}
