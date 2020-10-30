package dmn

import (
	"log"
	"os"
	"testing"
)

func TestListHandler(t *testing.T) {

	err := InitHandlerTest()

	if err != nil {
		t.Errorf("Error initializing test %v", err)
	}

	// Create a command
	var cmd Command
	cmd.Set("ls", "list files")

	// Manually populate our history file
	var requestHandler RequestHandler

	requestHandler.Set(TestSecret, TestHistory)
	requestHandler.Log = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	ret := requestHandler.SaveCmd(cmd)

	if ret != true {
		t.Errorf("Unable to save command")
	}

	cmds, err := requestHandler.ListCmd()

	if err != nil {
		t.Errorf("Error when listing commands")
	}

	if len(cmds) != 1 {
		t.Errorf("No commands to list")
	}
}
