package dmn

import (
	"testing"
)

func TestRunHandler(t *testing.T) {

	var app App

	err := app.InitHandlerTest()

	if err != nil {
		t.Errorf("Error initializing test %v", err)
	}

	// Create a command
	var cmd Command
	cmd.Set("ls", "list files", ".")

	// Manually populate our history file
	// var requestHandler RequestHandler

	// requestHandler.Set(TestSecret, TestHistory)
	// requestHandler.Log = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	ret := app.RequestHandler.SaveCmd(cmd)

	if ret != true {
		t.Errorf("Unable to save command")
	}

	app.RequestHandler.CommandScheduler.CreateScheduler()

	go func() {
		app.RequestHandler.CommandScheduler.CommandQueue <- cmd
		app.RequestHandler.CommandScheduler.RunSchedulerMock()
	}()
}
