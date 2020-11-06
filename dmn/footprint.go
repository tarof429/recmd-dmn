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
			log.Fatalf("Error, unable to create testdata: %v\n", err)
		}
	}

	f.confDirPath = testdataDir
	f.binDirPath = testdataDir
	f.logDirPath = testdataDir
}

// DefaultFootprint creates a footprint for production. There are
// separate directories for configuration, binary files, and logs.
func (f *Footprint) DefaultFootprint() {
	wd, err := os.Getwd()

	if err != nil {
		log.Println(err)
	}

	parentDir := filepath.Dir(wd)

	f.confDirPath = filepath.Join(parentDir, "conf")
	f.binDirPath = filepath.Join(parentDir, "bin")
	f.logDirPath = filepath.Join(parentDir, "logs")

	for _, dir := range []string{f.confDirPath, f.binDirPath, f.logDirPath} {
		if os.IsNotExist(err) {
			mode := int(0755)
			err := os.Mkdir(dir, os.FileMode(mode))

			if err != nil {
				log.Fatalf("Error, unable to create dir: %v\n", dir)
			}
		}
	}
}
