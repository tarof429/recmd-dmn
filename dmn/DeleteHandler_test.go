package dmn

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

func TestDeleteHandler(t *testing.T) {

	InitHandlerTest()

	// Manually populate our history file
	var cmd Command
	cmd.Set("ls", "list files")
	var cmds []Command
	cmds = append(cmds, cmd)
	mode := int(0644)

	updatedData, _ := json.MarshalIndent(cmds, "", "\t")

	err := ioutil.WriteFile(TestHistory.Path, updatedData, os.FileMode(mode))
	if err != nil {
		t.Error(err)
	}

	var deleteHandler DeleteHandler

	deleteHandler.Set(TestSecret, TestHistory)

	ret, err := deleteHandler.DeleteCmd(cmd.CmdHash)

	if err != nil {
		t.Error(err)
	}

	if len(ret) != 1 {
		t.Errorf("Unable to delete command")
	}
}
