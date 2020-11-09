package dmn

import (
	"testing"
)

func TestRunHandler(t *testing.T) {

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

	app.CommandScheduler.CreateScheduler()

	go func() {
		app.CommandScheduler.CommandQueue <- cmd
		app.CommandScheduler.RunSchedulerMock()
	}()
}
