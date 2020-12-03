package main

import (
	"bytes"
	"fmt"
	"github.com/gookit/color"
	"github.com/operatios/lsg/category"
)

var (
	//red       = color.New(color.FgRed).SprintfFunc()
	//hiGreen   = color.New(color.FgHiGreen).SprintfFunc()
	//hiYellow  = color.New(color.FgHiYellow).SprintfFunc()
	//yellow    = color.YellowString
	//hiCyan    = color.HiCyanString
	//cyan      = color.CyanString
	//hiMagenta = color.HiMagentaString
	//magenta   = color.MagentaString

	cachedColor = map[string]*color.RGBColor{}

	Dark = &Theme{
		cs: &cs{
			mc: nil,
			oc: "#0426a8",
			gc: "#2c2d2c",
			nc: "#2c2c2c",
			sc: map[int]string{
				0:   "#efaa45", // others
				150: "#a66321", // >= 150MiB
				500: "#a22815", // >= 500MiB
			},
			tc: "#70cbdc",
			ec: map[int]string{
				category.File:       "#398424",
				category.Dir:        "#0426a8",
				category.Symlink:    "#2c2c2c",
				category.Broken:     "#388425",
				category.Archive:    "#50933e",
				category.Executable: "#78fa53",
				category.Code:       "#388425",
				category.Image:      "#eeab46",
				category.Audio:      "#eeab46",
				category.Video:      "#eeab46",
			},
		},
	}
	Light = &Theme{
		cs: &cs{
			mc: map[rune]string{
				'r': "#a56361",
				'w': "#b73931",
				'x': "#0326a8",
				'-': "#2b2c2c",
				'd': "#0426a8",
			},
			oc: "#0426a8",
			gc: "#2c2d2c",
			nc: "#2c2c2c",
			sc: map[int]string{
				0:   "#efaa45", // others
				150: "#a66321", // >= 150MiB
				500: "#a22815", // >= 500MiB
			},
			tc: "#70cbdc",
			ec: map[int]string{
				category.File:       "#398424",
				category.Dir:        "#0426a8",
				category.Symlink:    "#2c2c2c",
				category.Broken:     "#388425",
				category.Archive:    "#50933e",
				category.Executable: "#78fa53",
				category.Code:       "#388425",
				category.Image:      "#eeab46",
				category.Audio:      "#eeab46",
				category.Video:      "#eeab46",
			},
		},
	}
)

// Theme color definition
type cs struct {
	mc map[rune]string // mode color
	oc string          // owner color
	gc string          // group color
	nc string          // nLink color
	sc map[int]string  // size color
	tc string          // time color
	ec map[int]string  // entry color
}

// interface of Theme
type Theme struct {
	*cs
}

func (t *Theme) mode(format, mode string, align int) string {
	mode = fmt.Sprintf(format, align, mode)
	buffer := bytes.Buffer{}
	for _, c := range mode {
		buffer.WriteString(getColor(t.mc[c]).Sprint(string(c)))
	}
	return buffer.String()
}
func (t *Theme) nLink(format string, v ...interface{}) string {
	return color.FgBlack.Sprintf(format, v...)
}
func (t *Theme) owner(format string, v ...interface{}) string {
	return getColor(t.oc).Sprintf(format, v...)
}
func (t *Theme) group(format string, v ...interface{}) string {
	return getColor(t.gc).Sprintf(format, v...)
}
func (t *Theme) size(size int, format string, v ...interface{}) string {
	return getColor(t.sc[size]).Sprintf(format, v...)
}
func (t *Theme) time(format string, v ...interface{}) string {
	return ""
}
func (t *Theme) entry(file File, args Args) string {
	return ""
}
func (t *Theme) total(format string, v ...interface{}) string {
	return getColor(t.ec[category.File]).Sprintf(format, v...)
}

func getColor(hex string) *color.RGBColor {
	if c, ok := cachedColor[hex]; ok {
		return c
	}

	rgbColor := color.HEX(hex)
	cachedColor[hex] = &rgbColor
	return &rgbColor
}
