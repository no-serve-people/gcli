package helper

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/gookit/goutil/strutil"
)

const (
	// match a good option, argument name
	regGoodName = `^[a-zA-Z][\w-]*$`
	// match a good command name
	regGoodCmdName = `^[a-zA-Z][\w-]*$`
	// match command id. eg: "self:init"
	regGoodCmdId = `^[a-zA-Z][\w:-]*$`
	// match command path. eg: "self init"
	// regGoodCmdPath = `^[a-zA-Z][\w -]*$`
)

var (
	// GoodName good name for option and argument
	goodName = regexp.MustCompile(regGoodName)
	// GoodCmdId match a good command name
	goodCmdId = regexp.MustCompile(regGoodCmdId)
	// GoodCmdName match a good command name
	goodCmdName = regexp.MustCompile(regGoodCmdName)
)

func IsGoodName(name string) bool {
	return goodName.MatchString(name)
}

func IsGoodCmdName(name string) bool {
	return goodCmdName.MatchString(name)
}

func IsGoodCmdId(name string) bool {
	return goodCmdId.MatchString(name)
}

// exec: `stty -a 2>&1`
// const (
// mac: speed 9600 baud; 97 rows; 362 columns;
// macSttyMsgPattern = `(\d+)\s+rows;\s*(\d+)\s+columns;`
// linux: speed 38400 baud; rows 97; columns 362; line = 0;
// linuxSttyMsgPattern = `rows\s+(\d+);\s*columns\s+(\d+);`
// )

var (
	terminalWidth, terminalHeight int

	// macSttyMsgMatch = regexp.MustCompile(macSttyMsgPattern)
	// linuxSttyMsgMatch = regexp.MustCompile(linuxSttyMsgPattern)
)

// Panicf message
func Panicf(format string, v ...any) {
	panic(fmt.Sprintf("GCli: "+format, v...))
}

// RenderText render text template with data. TODO use strutil.RenderText()
func RenderText(input string, data any, fns template.FuncMap, isFile ...bool) string {
	t := template.New("cli")
	t.Funcs(template.FuncMap{
		// don't escape content
		"raw": func(s string) string {
			return s
		},
		"trim": strings.TrimSpace,
		// join strings. usage {{ join .Strings ","}}
		"join": func(ss []string, sep string) string {
			return strings.Join(ss, sep)
		},
		// lower first char
		"lcFirst": strutil.LowerFirst,
		// upper first char
		"ucFirst": strutil.UpperFirst,
	})

	// custom add template functions
	if len(fns) > 0 {
		t.Funcs(fns)
	}

	if len(isFile) > 0 && isFile[0] {
		template.Must(t.ParseFiles(input))
	} else {
		template.Must(t.Parse(input))
	}

	// use buffer receive rendered content
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		panic(err)
	}

	return buf.String()
}
