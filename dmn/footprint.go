package dmn

import (
	"log"
	"os"
	"path/filepath"
)

// Footprint rerepresents the layout of the application.
type Footprint struct {
	confDirPath string
	binDirPath  string
	logDirPath  string
}

// Set sets some global variables associated with the footprint
func (f *Footprint) Set(footprint *Footprint) {
	f.confDirPath = footprint.confDirPath
	f.binDirPath = footprint.binDirPath
	f.logDirPath = footprint.logDirPath
}

// TestFootprint creates a footprint for testing. All files will be put in a
// directory called 'testdata'.
func (f *Footprint) TestFootprint() {
	wd, err := os.Getwd()

	if err != nil {
		log.Println(err)
	}

	testdataDir := filepath.Join(wd, "testdata")

	_, err = os.Stat(testdataDir)

	if os.IsNotExist(err) {
		mode := int(0755)
		err := os.Mkdir(testdataDir, os.FileMode(mode))

		if err != nil {
			log.Fatalf("Error: unable to create testdata: %v\n", err)
		}
	}

	f.confDirPath = testdataDir
	f.binDirPath = testdataDir
	f.logDirPath = testdataDir
}

// DefaultFootprint creates a footprint for production. There are
// separate directories for configuration, binary files, and logs.
// The code only creastes the conf and logs directory because without either
// the rest of the code will have problems.
func (f *Footprint) DefaultFootprint() {

	wd, err := os.Getwd()

	if err != nil {
		log.Println(err)
	}

	f.confDirPath = filepath.Join(wd, "conf")
	f.binDirPath = filepath.Join(wd, "bin")
	f.logDirPath = filepath.Join(wd, "logs")

	for _, dir := range []string{f.confDirPath, f.logDirPath} {
		mode := int(0755)
		os.Mkdir(dir, os.FileMode(mode))
	}
}
