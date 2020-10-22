package dmn

import (
	"encoding/base64"
	"fmt"
	"os"
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

func TestCreateSecret(t *testing.T) {
	secret := new(Secret)
	secret.CreateSecret()
	secret.SetPathToSecretsFile("./testdata")
	err := secret.WriteSecretToFile()

	if err != nil {
		t.Error("Unable to write secret")
	}

	if "" == secret.GetSecret() {
		t.Error("Invalid secret")
	}

}
