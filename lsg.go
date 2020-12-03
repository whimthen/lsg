package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/bmatcuk/doublestar/v2"
)

func processGlob(path string, args Args) {
	fileNames := Glob(path)

	parents := make(map[string][]string)
	for _, fileName := range fileNames {
		dir := filepath.Dir(fileName)
		parents[dir] = append(parents[dir], fileName)
	}

	keys := make([]string, 0, len(parents))
	for k := range parents {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return cmpCaseInsensitive(keys[i], keys[j])
	})

	for _, parent := range keys {
		if !args.all && isPathHidden(parent) {
			continue
		}

		children := getParentFiles(parents[parent], args.all)
		if len(children) == 0 {
			continue
		}

		_, _ = fmt.Fprintf(bufStdout, "%s:\n", parent)
		processFiles(children, args)
	}
}

func processFiles(files []File, args Args) {
	sortFiles(files, args.sort, args.reverse)

	if args.longList {
		formatList(files, args)
	} else {
		formatGrid(files, args)
	}
}

func processTree(files []File, fromDepths map[int]bool, args Args) {
	if len(files) == 0 {
		return
	}

	sortFiles(files, args.sort, args.reverse)
	depth := len(splitPath(files[0].path)) - 1

	for _, file := range files {
		isLast := file == files[len(files)-1]

		if file.isDir() {
			if isLast {
				delete(fromDepths, depth)
			} else {
				fromDepths[depth] = true
			}
		}

		var prefix string
		for i := 0; i < depth; i++ {
			if exists := fromDepths[i]; exists {
				prefix += "│  "
			} else {
				prefix += "   "
			}
		}
		if isLast {
			prefix += "└──"
		} else {
			prefix += "├──"
		}

		_, _ = fmt.Fprintln(bufStdout, prefix+file.colored(args))

		if file.isDir() && !file.isLink() {
			subFiles, _ := getFiles(file.path, args.all)
			processTree(subFiles, fromDepths, args)
		}
	}
}

func Glob(pattern string) []string {
	var matches []string

	if !strings.Contains(pattern, "**") {
		matches, _ = filepath.Glob(pattern)
	} else {
		matches, _ = doublestar.Glob(pattern)
	}

	return matches
}

func getFiles(path string, showHidden bool) ([]File, error) {
	var result []File

	fileInfos, err := ioutil.ReadDir(path)

	if err != nil {
		return nil, err
	}

	for _, fileInfo := range fileInfos {
		file := File{fileInfo, filepath.Join(path, fileInfo.Name())}

		if showHidden || !file.isHidden() {
			result = append(result, file)
		}
	}
	return result, nil
}

func getParentFiles(fileNames []string, showHidden bool) []File {
	var result []File

	for _, fileName := range fileNames {
		file, err := newFile(fileName)

		if err != nil {
			continue
		}

		if showHidden || !file.isHidden() {
			result = append(result, file)
		}
	}
	return result
}

func getRowCol(i int, rows int) (int, int) {
	row := i % rows
	return row, (i - row) / rows
}

func formatRows(files []File, columns int, args Args) [][]string {
	var rows int
	if len(files)%columns != 0 {
		rows = (len(files) / columns) + 1
	} else {
		rows = len(files) / columns
	}

	rowSlice := make([][]string, rows)
	columnWidths := make([]int, columns)

	for i, file := range files {
		_, col := getRowCol(i, rows)

		nameLength := utf8.RuneCountInString(file.pretty(args))
		if nameLength > columnWidths[col] {
			columnWidths[col] = nameLength
		}
	}

	var rowWidth int
	for _, width := range columnWidths {
		rowWidth += width
	}
	if rowWidth+(len(columnWidths)-1)*args.colSep >= terminalWidth {
		return nil
	}

	for i, file := range files {
		row, col := getRowCol(i, rows)
		wsAmt := columnWidths[col] - utf8.RuneCountInString(file.pretty(args))
		padding := strings.Repeat(" ", wsAmt)

		rowSlice[row] = append(rowSlice[row], file.colored(args)+padding)
	}
	return rowSlice
}

func formatGrid(files []File, args Args) {
	columns := 2
	goingBackwards := false

	if args.columns > 0 {
		columns = args.columns
		goingBackwards = true
	}

	var rows [][]string
	for columns > 1 {
		rows = formatRows(files, columns, args)
		if goingBackwards && rows != nil {
			break
		}

		if rows == nil || columns > len(files) {
			goingBackwards = true
		}

		if !goingBackwards {
			columns *= 2
		} else {
			columns--
		}
	}

	if columns > 1 {
		for i := range rows {
			sep := strings.Repeat(" ", args.colSep)
			_, _ = fmt.Fprintln(bufStdout, strings.Join(rows[i], sep))
		}
	} else {
		for i := range files {
			_, _ = fmt.Fprintln(bufStdout, files[i].colored(args))
		}
	}
}

func formatList(files []File, args Args) {
	sizes := []int64{}
	var totalSize int64

	var align struct {
		size     int
		fileMode int
		nLink    int
		owner    int
		group    int
	}

	theme := Light

	for _, file := range files {
		var sizeEntry string
		totalSize += file.size()

		if args.bytes {
			sizeEntry = fmt.Sprintf("%d %c", file.size(), 'B')
		} else {
			sizeEntry = humanizeSize(file.size())
		}
		sizes = append(sizes, file.size())

		if len(sizeEntry) > align.size {
			align.size = len(sizeEntry)
		}

		if args.listExtend {
			modeLen := len(file.fileMode())
			if modeLen > align.fileMode {
				align.fileMode = modeLen
			}

			nLinkLen := len(fmt.Sprint(file.nLink()))
			if nLinkLen > align.nLink {
				align.nLink = nLinkLen
			}
		}

		if args.listExtend && runtime.GOOS != "windows" {
			ownerLen := len(file.owner())
			if ownerLen > align.owner {
				align.owner = ownerLen
			}
			groupLen := len(file.group())
			if groupLen > align.group {
				align.group = groupLen
			}
		}
	}

	if args.bytes {
		_, _ = fmt.Fprintf(bufStdout, theme.total(args, "total %s\n", strconv.FormatInt(totalSize, 10)))
	} else {
		_, _ = fmt.Fprintf(bufStdout, theme.total(args, "total %s\n", humanizeSize(totalSize)))
	}

	for i, file := range files {
		var line string
		if args.listExtend {
			line += theme.mode(args, "%-*s   ", file.fileMode(), align.fileMode)
			line += theme.nLink(args, "%*d  ", align.nLink, file.nLink())
		}

		if args.listExtend && runtime.GOOS != "windows" {
			owner := file.owner()
			group := file.group()

			// WSL: file owner of /mnt/ is ""
			if owner == "" {
				owner = group
			}

			line += theme.owner(args, "%-*s  ", align.owner, owner)
			line += theme.group(args, "%-*s", align.group, group)
		}

		//sizeEntry := fmt.Sprintf("%*s", align.size+3, sizes[i])
		//if !args.noColors {
		//	sizeEntry = aurora.Colorize(sizeEntry, aurora.GreenFg).String()
		//}
		line += theme.size(args, "%*s", sizes[i], align.size+3)
		line += theme.time(args, file, 3)
		//line += files[i].colored(args)
		line += theme.entry(args, files[i])

		_, _ = fmt.Fprintln(bufStdout, line)
	}
}
