package main

import (
	"bytes"
	"fmt"
	"github.com/gookit/color"
	"github.com/operatios/lsg/category"
	"github.com/operatios/lsg/icons"
	"strings"
)

const (
	MB = 1024 * 1024
)

var (
	eeab46  = color.NewRGBStyle(color.HEX("#B8860B"))
	c388425 = color.NewRGBStyle(color.HEX("#388425"))
	c0426a8 = color.HEX("#0426a8")
	link    = color.HEX("#4169E1")
	Dark    = &Theme{
		oc:  color.HEX("#fffedb"),
		gc:  color.HEX("#d7d691"),
		nc:  color.HEX("#ffffff"),
		tc:  color.HEX("#71ad8a"),
		lc:  color.HEX("#eb6b34"),
		orc: color.FgLightRed,
		mc: map[rune]color.RGBColor{
			'r': color.HEX("#7ed36e"),
			'w': color.HEX("#d7d691"),
			'x': color.HEX("#b73831"),
			'-': color.HEX("#cd8b89"),
			'd': color.HEX("#4aaef8"),
			'L': link,
		},
		sc: map[int]color.RGBColor{
			0:    color.HEX("#fffedb"), // others
			150:  color.HEX("#fffa53"), // >= 150MiB
			500:  color.HEX("#f4b13e"), // >= 500MiB
			1024: color.HEX("#CD950C"), // >= 1G
		},
		ec: map[int]*color.RGBStyle{
			category.File:       color.NewRGBStyle(color.HEX("#6ff44a")),
			category.Dir:        color.NewRGBStyle(color.HEX("#4aaef8")),
			category.Symlink:    color.NewRGBStyle(color.HEX("#ebb434")),
			category.Archive:    color.NewRGBStyle(color.HEX("#cd0000")).AddOpts(color.OpUnderscore),
			category.Executable: color.NewRGBStyle(color.HEX("#78fa53")),
			category.Broken:     color.NewRGBStyle(color.HEX("#eb3434")),
			category.Code:       c388425,
			category.Image:      eeab46,
			category.Audio:      eeab46,
			category.Video:      eeab46,
		},
	}
	Light = &Theme{
		oc:  color.HEX("#191970"),
		gc:  color.HEX("#808000"),
		nc:  color.HEX("#2c2c2c"),
		tc:  color.HEX("#4682B4"),
		lc:  color.HEX("#225db5"),
		orc: color.FgRed,
		mc: map[rune]color.RGBColor{
			'r': color.HEX("#a56361"),
			'w': color.HEX("#b73931"),
			'x': color.HEX("#0326a8"),
			'-': color.HEX("#2b2c2c"),
			'd': c0426a8,
			'L': link,
		},
		sc: map[int]color.RGBColor{
			0:    color.HEX("#efaa45"), // others
			150:  color.HEX("#a66321"), // >= 150MiB
			500:  color.HEX("#a22815"), // >= 500MiB
			1024: color.HEX("#8B008B"), // >= 1G
		},
		ec: map[int]*color.RGBStyle{
			category.File:       color.HEXStyle("#228B22"),
			category.Dir:        color.NewRGBStyle(c0426a8),
			category.Symlink:    color.NewRGBStyle(link),
			category.Archive:    color.HEXStyle("#cd0000").AddOpts(color.OpUnderscore),
			category.Executable: color.HEXStyle("#006400"),
			category.Broken:     color.HEXStyle("#CD2626").AddOpts(color.OpBold),
			category.Code:       c388425,
			category.Image:      eeab46,
			category.Audio:      eeab46,
			category.Video:      eeab46,
		},
	}
)

// Theme color definition
type Theme struct {
	mc  map[rune]color.RGBColor // mode color
	oc  color.RGBColor          // owner color
	gc  color.RGBColor          // group color
	nc  color.RGBColor          // nLink color
	sc  map[int]color.RGBColor  // size color
	tc  color.RGBColor          // time color
	ec  map[int]*color.RGBStyle // entry color
	orc color.Color             // owner root color
	lc  color.RGBColor          // link real color
}

func (t *Theme) mode(args Args, format, mode string, align int) string {
	mode = fmt.Sprintf(format, align, mode)
	if args.noColors {
		return mode
	}
	buffer := bytes.Buffer{}
	for _, c := range mode {
		buffer.WriteString(t.mc[c].Sprint(string(c)))
	}
	return buffer.String()
}

func (t *Theme) nLink(args Args, format string, v ...interface{}) string {
	if args.noColors {
		return fmt.Sprintf(format, v...)
	}
	return color.FgDefault.Sprintf(format, v...)
}

func (t *Theme) owner(args Args, format, owner string, align int) string {
	if args.noColors {
		return fmt.Sprintf(format, align, owner)
	}
	if owner == "root" {
		return t.orc.Sprintf(format, align, owner)
	}
	return t.oc.Sprintf(format, align, owner)
}

func (t *Theme) group(args Args, format string, v ...interface{}) string {
	if args.noColors {
		return fmt.Sprintf(format, v...)
	}
	return t.gc.Sprintf(format, v...)
}

func (t *Theme) size(args Args, format string, size int64, align int) string {
	if args.noColors {
		return fmt.Sprintf(format, align, humanizeSize(size))
	}
	colorKey := 0
	if size > 1024 {
		if size >= 150*MB {
			if size < 500*MB {
				colorKey = 150
			} else if size < 1024*MB {
				colorKey = 500
			} else {
				colorKey = 1024
			}
		}
	}
	return t.sc[colorKey].Sprintf(format, align, humanizeSize(size))
}

func (t *Theme) time(args Args, f File, alignOffset int) string {
	formatted := f.info.ModTime().Format("Mon Jan 02 15:04:05 2006")
	if args.noColors {
		return fmt.Sprintf("%*s  ", len(formatted)+alignOffset, formatted)
	}
	return t.tc.Sprintf("%*s  ", len(formatted)+alignOffset, formatted)
}

func (t *Theme) entry(args Args, f File) string {
	pretty := f.pretty(args)

	if args.noColors {
		return pretty
	}

	if f.isBroken() {
		pretty += " [Dead link]"
	}
	if f.isLink() {
		arrowIndex := strings.Index(pretty, icons.LinkArrow)
		link := pretty[:arrowIndex]
		realf := pretty[arrowIndex:]
		return t.ec[f.category()].Sprint(link) + t.lc.Sprint(realf)
	}
	return t.ec[f.category()].Sprint(pretty)
}

func (t *Theme) total(args Args, format string, v ...interface{}) string {
	if args.noColors {
		return fmt.Sprintf(format, v...)
	}
	return t.ec[category.File].Sprintf(format, v...)
}
