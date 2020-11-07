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

type LogFile struct {
	Path string
	Log  *log.Logger
}

func (l *LogFile) Set(path string) {
	l.Path = filepath.Join(path, logsFile)
}

func (l *LogFile) Create() {

	f, err := os.Create(l.Path)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	l.Log = log.New(f, "", log.LstdFlags|log.Lshortfile)

}
