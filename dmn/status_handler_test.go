package dmn

import (
	"testing"
)

func TestStatusHandler(t *testing.T) {

	var app App

	err := app.InitalizeTest()

	if err != nil {
		t.Errorf("Error initializing test %v", err)
	}

	ret, err := app.StatusCmd()

	if err != nil {
		t.Errorf("Unable to select command")
	}

	if ret != true {
		t.Errorf("Unable to get status")
	}

}
