package dmn

import (
	"encoding/base64"
)

// RequestVariable represents variables passed into the request
type RequestVariable struct {
	Secret           string
	Description      string
	CmdHash          string
	Command          string
	WorkingDirectory string
}

// GetVariablesFromRequestVars gets variables from the request
func (variables *RequestVariable) GetVariablesFromRequestVars(vars map[string]string) error {

	secret, err := base64.StdEncoding.DecodeString(vars["secret"])

	if err != nil {
		return err
	}

	cmdHash, err := base64.StdEncoding.DecodeString(vars["cmdHash"])

	if err != nil {
		return err
	}

	description, err := base64.StdEncoding.DecodeString(vars["description"])
	if err != nil {
		return err
	}

	command, err := base64.StdEncoding.DecodeString(vars["command"])
	if err != nil {
		return err
	}

	workingDirectory, err := base64.StdEncoding.DecodeString(vars["workingDirectory"])
	if err != nil {
		return err
	}

	variables.Secret = string(secret)
	variables.CmdHash = string(cmdHash)
	variables.Description = string(description)
	variables.Command = string(command)
	variables.WorkingDirectory = string(workingDirectory)

	return nil
}
