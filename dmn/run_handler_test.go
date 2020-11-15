package dmn

import (
	"fmt"
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
	cmd.Set("ls", "list files", "testdata")

	ret := app.SaveCmd(cmd)

	if ret != true {
		t.Errorf("Unable to save command")
	}

	app.CreateScheduler()

	go func() {
		app.RunScheduler()
		fmt.Println("After RunScheduler")

	}()

	app.CommandScheduler.CommandQueue <- cmd

	sc := <-app.CommandScheduler.CompletedQueue
	fmt.Printf("Status: %v\n", sc.Status)
	fmt.Printf("Output: %v\n", sc.Coutput)

}

func TestRunCmdWithBadWorkingDirectory(t *testing.T) {

	var app App

	err := app.InitalizeTest()

	if err != nil {
		t.Errorf("Error initializing test %v", err)
	}

	var cmd Command

	cmd.Set("another-command", "another-command", ".")

	ret := app.SaveCmd(cmd)

	if ret != true {
		t.Errorf("Unable to save command")
	}

	cmd.WorkingDirectory = "bad directory"

	fmt.Println("Creating scheduler")
	app.CreateScheduler()

	go func() {
		app.RunSchedulerMock(Failed)
		fmt.Println("After RunSchedulerMock")

	}()

	app.CommandScheduler.CommandQueue <- cmd

	//time.Sleep(time.Second * 3)

	sc := <-app.CommandScheduler.CompletedQueue
	fmt.Printf("Status: %v\n", sc.Status)
}
