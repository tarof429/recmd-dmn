package dmn

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	//dmn "github.com/tarof429/recmd-dmn/dmn"
)

const testdataDir = "testdata"
const testHistoryFile = testdataDir + "/.cmd_history.json"

func TestMain(m *testing.M) {
	fmt.Println("Running tests...")

	status := m.Run()

	os.Exit(status)
}

func getBase64(line string) string {

	lineData := []byte(line)
	return base64.StdEncoding.EncodeToString(lineData)
}

// Test whether getVariablesFromRequestVars decodes base 64 strings
func TestGetVariablesFromRequestVars(t *testing.T) {

	secret := "GBrOoGB5F5PJ1kJ9oA4qs7FCEmMp7IID1E6NLtAB"
	description := "Calculate disk size"
	cmdHash := "4048dc4ddd288eac474c7550bc632c"

	var variables RequestVariable

	var vars = make(map[string]string)
	vars["secret"] = getBase64(secret)
	vars["description"] = getBase64(description)
	vars["cmdHash"] = getBase64(cmdHash)

	err := variables.GetVariablesFromRequestVars(vars)

	if err != nil {
		t.Error("Unable to get variables from request vars")
	}

	if description != variables.Description {
		t.Error("Description did not match")
	}
}

func TestDeleteHandler(t *testing.T) {

	var (
		history HistoryFile
		secret  Secret
	)

	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	// Write the secret
	secret.Set(filepath.Join(path, "testdata"))
	err = secret.WriteSecretToFile()
	if err != nil {
		t.Error(err)
	}
	if secret.GetSecret() == "" {
		t.Errorf("Secret was an empty string")
	}

	// Create a history file
	history.Set(filepath.Join(path, "testdata"))
	err = history.WriteHistoryToFile()
	if err != nil {
		t.Error(err)
	}

	// Manually populate our history file
	var cmd Command
	cmd.Set("ls", "list files")
	var cmds []Command
	cmds = append(cmds, cmd)
	mode := int(0644)

	updatedData, _ := json.MarshalIndent(cmds, "", "\t")

	err = ioutil.WriteFile(history.Path, updatedData, os.FileMode(mode))
	if err != nil {
		t.Error(err)
	}

	var deleteHandler DeleteHandler

	deleteHandler.Set(secret, history)

	ret, err := deleteHandler.DeleteCmd(cmd.CmdHash)

	if err != nil {
		t.Error(err)
	}

	if len(ret) != 1 {
		t.Errorf("Unable to delete command")
	}
}
