package dmn

import (
	"fmt"
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
			log.Fatalf("Error: unable to create testdata: %v\n", err)
		}
	}

	f.confDirPath = testdataDir
	f.binDirPath = testdataDir
	f.logDirPath = testdataDir
}

// DefaultFootprint creates a footprint for production. There are
// separate directories for configuration, binary files, and logs.
func (f *Footprint) DefaultFootprint() {
	fmt.Println("Creating footprint")
	log.Printf("Creating footprint")

	wd, err := os.Getwd()

	if err != nil {
		log.Println(err)
	}

	parentDir := filepath.Dir(wd)

	f.confDirPath = filepath.Join(parentDir, "conf")
	f.binDirPath = filepath.Join(parentDir, "bin")
	f.logDirPath = filepath.Join(parentDir, "logs")

	//os.Mkdir(f.confDirPath, 0755)

	// f.confDirPath = filepath.Join(wd, "conf")
	// f.binDirPath = filepath.Join(wd, "bin")
	// f.logDirPath = filepath.Join(wd, "logs")

	for _, dir := range []string{f.confDirPath, f.logDirPath} {
		if os.IsNotExist(err) {
			fmt.Println("Skipping...")

		} else {
			fmt.Printf("Creating %v\n", dir)
			mode := int(0755)
			err := os.Mkdir(dir, os.FileMode(mode))

			if err != nil {
				log.Printf("Warning: unable to create dir: %v\n", err)
			}
		}
	}
}
