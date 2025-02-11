// Package gflag provide command line options and arguments binding, parse, management.
package gflag

import (
	"github.com/gookit/goutil/cflag"
	"github.com/gookit/goutil/strutil"
)

const (
	// AlignLeft Align right, padding left
	AlignLeft = strutil.PosRight
	// AlignRight Align left, padding right
	AlignRight = strutil.PosLeft

	// default desc
	defaultDesc = "No description"

	// TagRuleNamed struct tag use named k-v rule.
	//
	// eg: `flag:"name=int0;shorts=i;required=true;desc=int option message"`
	TagRuleNamed = 0

	// TagRuleSimple struct tag use simple rule.
	// format: "desc;required;default;shorts"
	//
	// eg: `flag:"int option message;required;;i"`
	TagRuleSimple = 1
)

// FlagTagName default tag name on struct
var FlagTagName = "flag"

// Config for render help information
type Config struct {
	// WithoutType don't display flag data type on print help
	WithoutType bool
	// DescNewline flag desc at new line on print help
	DescNewline bool
	// Alignment flag name align left or right. default is: left
	Alignment strutil.PosFlag
	// TagName on struct
	TagName string
	// TagRuleType for struct tag value. default is TagRuleNamed
	TagRuleType uint8
	// DisableArg disable binding arguments.
	DisableArg bool
}

// OptCategory struct
type OptCategory struct {
	Name, Title string
	OptNames    []string
}

// Ints The int flag list, implemented flag.Value interface
type Ints = cflag.Ints

// Strings The string flag list, implemented flag.Value interface
type Strings = cflag.Strings

// Booleans The bool flag list, implemented flag.Value interface
type Booleans = cflag.Booleans

// EnumString The string flag list, implemented flag.Value interface
type EnumString = cflag.EnumString

// KVString The key-value string flag, repeatable.
type KVString = cflag.KVString

// ConfString The config-string flag, INI format, like nginx-config.
type ConfString = cflag.ConfString
