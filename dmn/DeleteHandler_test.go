package dmn

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestDeleteHandler(t *testing.T) {

	err := InitHandlerTest()

	if err != nil {
		t.Errorf("Error initializing test %v", err)
	}

	// Manually populate our history file
	var cmd Command
	cmd.Set("ls", "list files")
	var cmds []Command
	cmds = append(cmds, cmd)
	mode := int(0644)

	updatedData, _ := json.MarshalIndent(cmds, "", "\t")

	err = ioutil.WriteFile(TestHistory.Path, updatedData, os.FileMode(mode))
	if err != nil {
		t.Error(err)
	}

	var requestHandler RequestHandler

	requestHandler.Set(TestSecret, TestHistory)
	requestHandler.Log = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	ret, err := requestHandler.DeleteCmd(cmd.CmdHash)

	if err != nil {
		t.Error(err)
	}

	if len(ret) != 1 {
		t.Errorf("Unable to delete command")
	}
}
