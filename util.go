package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func isPathHidden(path string) bool {
	if path == "." {
		return false
	}

	components := splitPath(path)

	for i := 1; i <= len(components); i++ {
		f := filepath.Join(components[:i]...)

		if strings.Contains(f, "..") {
			abs, _ := filepath.Abs(f)
			f = abs
		}

		file, err := newFile(f)
		if err != nil {
			continue
		}

		if file.isHidden() {
			return true
		}
	}
	return false
}

func splitPath(path string) []string {
	return strings.Split(path, string(filepath.Separator))
}

func humanizeSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B  ", size)
	}

	fSize := float64(size)
	fSize /= 1024

	for _, unit := range []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB"} {
		if fSize < 9 {
			return fmt.Sprintf("%.1f %s", fSize, unit)
		} else if fSize < 1000 {
			return fmt.Sprintf("%.0f %s", fSize, unit)
		}
		fSize /= 1024
	}
	return fmt.Sprintf("%.1f %s", fSize, "YiB")
}
