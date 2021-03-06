package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	dmn "github.com/tarof429/recmd-dmn/dmn"
)

var a dmn.App

func TestMain(m *testing.M) {

	fmt.Println("Called TestMain")

	a.InitalizeTest()

	a.Router = mux.NewRouter()

	a.InitializeRoutes()

	a.CreateScheduler()
	go a.RunScheduler()
	go a.QueuedCommandsCleanup()

	code := m.Run()

	os.Exit(code)
}

func clearConfigDir() {

	os.RemoveAll(dmn.TestConfigDir)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {

	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder, req)

	return recorder
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func getBase64(line string) string {

	lineData := []byte(line)
	return base64.StdEncoding.EncodeToString(lineData)
}

func makeEndpoint(endpoint string, params map[string]string) string {

	for key, value := range params {
		endpoint = strings.Replace(endpoint, key, getBase64(value), -1)
	}

	return endpoint
}

func clearHistory() {
	a.History.Remove()
	a.History.Create()
	// os.Remove(filepath.Join(dmn.TestConfigDir, "recmd_history.json"))
	// os.Create(filepath.Join(dmn.TestConfigDir, "recmd_history.json"))
}

// func TestConfigDirExists(t *testing.T) {

// 	clearHistory()

// 	fileInfo, statErr := os.Stat(dmn.TestConfigDir)

// 	if os.IsNotExist(statErr) {
// 		t.Errorf("%v does not exist\n", dmn.TestConfigDir)
// 	} else if !fileInfo.IsDir() {
// 		t.Errorf("%v is not a directory", dmn.TestConfigDir)
// 	}
// }

func TestListHandler(t *testing.T) {

	clearHistory()

	// Add
	func() {
		endpoint := "/secret/{secret}/add/command/{command}/description/{description}/workingDirectory/{workingDirectory}"

		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()
		params["{command}"] = "ls -ltr /"
		params["{description}"] = "list files"
		params["{workingDirectory}"] = "."

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)

		var ret string

		json.Unmarshal(response.Body.Bytes(), &ret)

		if ret != "true" {
			t.Errorf("Unable to save command")
		}
	}()

	// list
	func() {
		endpoint := "/secret/{secret}/list"

		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)
	}()
}

func TestAddHandler(t *testing.T) {

	clearHistory()

	// Add
	func() {

		endpoint := "/secret/{secret}/add/command/{command}/description/{description}/workingDirectory/{workingDirectory}"

		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()
		params["{command}"] = "ls -ltr /"
		params["{description}"] = "list files"
		params["{workingDirectory}"] = "."

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)

		var ret string

		json.Unmarshal(response.Body.Bytes(), &ret)

		if ret != "true" {
			t.Errorf("Unable to save command")
		}
	}()

	// List
	func() {

		endpoint := "/secret/{secret}/list"

		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)

		var cmds []dmn.Command

		json.Unmarshal(response.Body.Bytes(), &cmds)

		if len(cmds) != 1 {
			t.Errorf("No commands found")
		}
		if cmds[0].CmdString != "ls -ltr /" {
			t.Errorf("Command string is invalid")
		}
	}()

}

func TestSearchHandler(t *testing.T) {

	clearHistory()

	add := func(cmd string) {
		endpoint := "/secret/{secret}/add/command/{command}/description/{description}/workingDirectory/{workingDirectory}"

		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()
		params["{command}"] = cmd
		params["{description}"] = "Dummy description"
		params["{workingDirectory}"] = "."

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)
	}

	for i := 0; i < 10; i++ {
		add("command " + strconv.Itoa(i))
	}

	search := func(description string, expectedCount int) {
		endpoint := "/secret/{secret}/search/description/{description}"

		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()
		params["{description}"] = description

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)

		var cmds []dmn.Command

		json.Unmarshal(response.Body.Bytes(), &cmds)

		if len(cmds) != expectedCount {
			t.Errorf("Not enough!")
		}

	}

	func() {
		desc := []string{"Dummy", "d", "DUMMY", "my"}

		for _, d := range desc {

			search(d, 10)
		}
	}()

	func() {
		desc := []string{"X", "12345", "/", "@$!"}

		for _, d := range desc {

			search(d, 0)
		}
	}()
}

func TestAdd2Items(t *testing.T) {

	clearHistory()

	add := func(command, description string) {
		// Add command
		endpoint := "/secret/{secret}/add/command/{command}/description/{description}/workingDirectory/{workingDirectory}"

		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()
		params["{command}"] = command
		params["{description}"] = description
		params["{workingDirectory}"] = "."

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)

		var ret string

		json.Unmarshal(response.Body.Bytes(), &ret)

		if ret != "true" {
			t.Errorf("Unable to save command")
		}
	}

	add("rm -f ", "delete files")
	add("df -h", "Check disk space")

}

func TestSelecthHandler(t *testing.T) {

	clearHistory()

	add := func() {
		// Add command
		endpoint := "/secret/{secret}/add/command/{command}/description/{description}/workingDirectory/{workingDirectory}"

		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()
		params["{command}"] = "uname -srm"
		params["{description}"] = "Check linux version"
		params["{workingDirectory}"] = "."

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)

		var ret string

		json.Unmarshal(response.Body.Bytes(), &ret)
		if ret != "true" {
			t.Errorf("Unable to save command")
		}
	}

	add()

	list := func() string {
		// Use list command to get the Command with the hash we want
		endpoint := "/secret/{secret}/list"
		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)

		var cmds []dmn.Command
		json.Unmarshal(response.Body.Bytes(), &cmds)

		if len(cmds) != 1 {
			t.Errorf("No commands found")
		}

		if cmds[0].CmdString != "uname -srm" {
			t.Errorf("Command string is invalid")
		}

		return cmds[0].CmdHash
	}

	expectedHash := list()

	fmt.Printf("Expected hash: %v\n", expectedHash)

	selectFunc := func(expectedHash string) {
		endpoint := "/secret/{secret}/select/cmdHash/{cmdHash}"
		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()
		params["{cmdHash}"] = expectedHash

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)

		var cmd dmn.Command
		json.Unmarshal(response.Body.Bytes(), &cmd)

		if cmd.CmdString != "uname -srm" {
			t.Errorf("Command string is invalid")
		}
	}

	selectFunc(expectedHash)
}

func TestDeleteHandlerPartial(t *testing.T) {

	clearHistory()

	add := func(cmd string) {
		endpoint := "/secret/{secret}/add/command/{command}/description/{description}/workingDirectory/{workingDirectory}"

		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()
		params["{command}"] = cmd
		params["{description}"] = "Dummy description"
		params["{workingDirectory}"] = "."

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)
	}

	for i := 0; i < 10; i++ {
		add("command " + strconv.Itoa(i))
	}

	list := func() []string {
		// Use list command to get the Command with the hash we want
		endpoint := "/secret/{secret}/list"
		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)

		var cmds []dmn.Command
		json.Unmarshal(response.Body.Bytes(), &cmds)

		var ret []string

		for _, cmd := range cmds {
			ret = append(ret, cmd.CmdHash)

		}
		return ret
	}

	founddHashes := list()

	// Leave the first command alone and delete the others
	founddHashes = founddHashes[1:]

	if len(founddHashes) != 9 {
		t.Errorf("We should only have 9 hashes!")
	}

	deleteFunc := func(cmdHash string) {
		endpoint := "/secret/{secret}/delete/cmdHash/{cmdHash}"

		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()
		params["{cmdHash}"] = cmdHash

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)
	}

	for _, hash := range founddHashes {
		deleteFunc(hash)
	}

	founddHashes = list()

	if len(founddHashes) != 1 {
		t.Errorf("Commands not deleted")
	}
}

func TestShowHandler(t *testing.T) {

	clearHistory()

	createFile := func() {
		pathToScript := filepath.Join("testdata", "test.sh")
		ioutil.WriteFile(pathToScript, []byte("#!/bin/sh\n# My script\n"), os.FileMode(0644))
	}

	createFile()

	add := func() {
		// Add command
		endpoint := "/secret/{secret}/add/command/{command}/description/{description}/workingDirectory/{workingDirectory}"

		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()
		params["{command}"] = "sh ./test.sh"
		params["{description}"] = "Run test.sh"
		params["{workingDirectory}"] = "testdata"

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)

		var ret string

		json.Unmarshal(response.Body.Bytes(), &ret)
		if ret != "true" {
			t.Errorf("Unable to save command")
		}
	}

	add()

	list := func() string {
		// Use list command to get the Command with the hash we want
		endpoint := "/secret/{secret}/list"
		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)

		var cmds []dmn.Command
		json.Unmarshal(response.Body.Bytes(), &cmds)

		if len(cmds) != 1 {
			t.Errorf("No commands found")
		}

		if cmds[0].CmdString != "sh ./test.sh" {
			t.Errorf("Command string is invalid")
		}

		return cmds[0].CmdHash
	}

	expectedHash := list()

	selectFunc := func(expectedHash string) {
		endpoint := "/secret/{secret}/show/cmdHash/{cmdHash}"
		params := make(map[string]string)
		params["{secret}"] = a.Secret.GetSecret()
		params["{cmdHash}"] = expectedHash

		endpoint = makeEndpoint(endpoint, params)

		req, _ := http.NewRequest("GET", endpoint, nil)

		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)

		var ret string
		json.Unmarshal(response.Body.Bytes(), &ret)

		fmt.Println(ret)
	}

	selectFunc(expectedHash)
}
