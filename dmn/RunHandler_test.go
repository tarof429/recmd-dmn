package dmn

import "testing"

func TestRunHandler(t *testing.T) {

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
	ret := requestHandler.SaveCmd(cmd)

	if ret != true {
		t.Errorf("Unable to save command")
	}

	var sc ScheduledCommand

	requestHandler.ScheduleCommand(cmd, sc.RunMockCommand)

	if sc.ExitStatus != 99 {
		t.Errorf("Command did not successfully run")
	}
}
