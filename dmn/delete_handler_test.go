package dmn

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

func TestDeleteHandler(t *testing.T) {

	var app App

	err := app.InitalizeTest()

	if err != nil {
		t.Errorf("Error initializing test %v", err)
	}

	// Manually populate our history file
	var cmd Command
	cmd.Set("ls", "list files", ".")
	var cmds []Command
	cmds = append(cmds, cmd)
	mode := int(0644)

	updatedData, _ := json.MarshalIndent(cmds, "", "\t")

	err = ioutil.WriteFile(app.History.Path, updatedData, os.FileMode(mode))
	if err != nil {
		t.Error(err)
	}

	ret, err := app.DeleteCmd(cmd.CmdHash)

	if err != nil {
		t.Error(err)
	}

	if len(ret) != 1 {
		t.Errorf("Unable to delete command")
	}
}
