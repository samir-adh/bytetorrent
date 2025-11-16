package log

import (
	"fmt"
	"log"
)

type Logger struct {
	Verbose VerboseLevel
}

func (logger *Logger) Printf(verboseLevel VerboseLevel, format string, v ...any) {
	if verboseLevel <= logger.Verbose {
		log.Printf(format, v...)
	}
}

func (logger *Logger) Println(verboseLevel VerboseLevel, v ...any) {
	if verboseLevel <= logger.Verbose {
		log.Println(v...)
	}
}
func (logger *Logger) Print(verboseLevel VerboseLevel, v ...any) {
	if verboseLevel <= logger.Verbose {
		log.Print(v...)
	}
}

func (logger *Logger) Progress(progress int) {
	if logger.Verbose == HighVerbose {
		return
	}
	progressBar := "["
	for i := range 50 {
		if 2*i < progress {
			progressBar += "-"
		} else {
			progressBar += " "
		}
	}
	progressBar += "]"

	fmt.Printf("\r%s Download %d%% complete", progressBar, progress)
	if progress == 100 {
		fmt.Println()
	}
}

type VerboseLevel int

const (
	LowVerbose  VerboseLevel = 0
	HighVerbose VerboseLevel = 1
)
