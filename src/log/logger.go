package log

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	Verbose       VerboseLevel
	progressMutex sync.Mutex
	startTime     time.Time
	lastProgress  int
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

// Progress displays an enhanced progress bar with colors and statistics
func (logger *Logger) Progress(progress int, downloaded, total int64) {
	if logger.Verbose == HighVerbose {
		return
	}

	logger.progressMutex.Lock()
	defer logger.progressMutex.Unlock()

	// Initialize start time on first call
	if logger.startTime.IsZero() {
		logger.startTime = time.Now()
	}

	// Clear the current line
	fmt.Print("\r\033[K")

	// Build progress bar with Unicode block characters
	barWidth := 40
	filled := int(float64(barWidth) * float64(progress) / 100.0)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	// Color codes
	const (
		colorReset  = "\033[0m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorCyan   = "\033[36m"
		colorBold   = "\033[1m"
	)

	// Calculate speed and ETA
	elapsed := time.Since(logger.startTime)
	var speedStr, etaStr string

	if downloaded > 0 && elapsed.Seconds() > 0 {
		speed := float64(downloaded) / elapsed.Seconds() // bytes per second
		speedStr = formatBytes(speed) + "/s"

		if progress > 0 {
			remaining := float64(total-downloaded) / speed
			etaStr = formatDuration(time.Duration(remaining * float64(time.Second)))
		}
	}

	// Choose color based on progress
	barColor := colorYellow
	if progress >= 100 {
		barColor = colorGreen
	}

	// Format file size info
	sizeInfo := fmt.Sprintf("%s / %s", formatBytes(float64(downloaded)), formatBytes(float64(total)))

	// Print the progress bar
	fmt.Print(barColor + colorBold + "[" + bar + "]" + colorReset)
	fmt.Printf(" %s%d%%%s", colorCyan, progress, colorReset)

	if speedStr != "" {
		fmt.Printf(" | %s", speedStr)
	}

	if etaStr != "" {
		fmt.Printf(" | ETA: %s", etaStr)
	}

	fmt.Printf(" | %s", sizeInfo)

	// Print newline when complete
	if progress >= 100 {
		fmt.Println()
		fmt.Printf("%s✓ Download complete!%s\n", colorGreen+colorBold, colorReset)
	}

	logger.lastProgress = progress
}

// ProgressSimple is a simpler version without speed/ETA (backward compatible)
func (logger *Logger) ProgressSimple(progress int) {
	if logger.Verbose == HighVerbose {
		return
	}

	logger.progressMutex.Lock()
	defer logger.progressMutex.Unlock()

	// Clear line
	fmt.Print("\r\033[K")

	// Build bar
	barWidth := 50
	filled := progress / 2
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	// Color
	color := "\033[33m" // Yellow
	if progress >= 100 {
		color = "\033[32m" // Green
	}

	fmt.Printf("%s[%s] %d%%%s", color, bar, progress, "\033[0m")

	if progress >= 100 {
		fmt.Println("\n\033[32m✓ Complete!\033[0m")
	}
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.0f B", bytes)
	}
	div, exp := float64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", bytes/div, "KMGTPE"[exp])
}

// formatDuration converts duration to human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

type VerboseLevel int

const (
	LowVerbose  VerboseLevel = 0
	HighVerbose VerboseLevel = 1
)
