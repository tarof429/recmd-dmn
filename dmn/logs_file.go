package dmn

import (
	"log"
	"os"
	"path/filepath"
)

const (
	// The log file
	logsFile = "recmd_dmn.log"
)

// LogFile represents the log file
type LogFile struct {
	Path string
	Log  *log.Logger
}

// Set sets the path to the log file
func (l *LogFile) Set(path string) {
	l.Path = filepath.Join(path, logsFile)
}

// Create creates the log file
func (l *LogFile) Create() {

	f, err := os.Create(l.Path)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	l.Log = log.New(f, "", log.LstdFlags|log.Lshortfile)

}
