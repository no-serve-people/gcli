package gcli

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/gookit/color"
	"github.com/gookit/gcli/v3/helper"
	"github.com/gookit/goutil/strutil"
)

// parseGlobalOpts parse global options
func (app *App) parseGlobalOpts(args []string) (ok bool) {
	Logf(VerbDebug, "will begin parse global options")

	// parse global options
	if !app.core.parseGlobalOpts(args) { // has error.
		return
	}

	// check global options
	if gOpts.showHelp {
		app.showApplicationHelp()
		return
	}

	if gOpts.showVer {
		app.showVersionInfo()
		return
	}

	// disable color
	if gOpts.NoColor {
		color.Enable = false
	}

	Debugf("global option parsed, verbose level is <mgb>%d</>", gOpts.verbose)
	app.args = app.GlobalFlags().FSetArgs()

	// TODO show auto-completion for bash/zsh
	if gOpts.inCompletion {
		app.showAutoCompletion(app.args)
		return
	}

	return true
}

// prepare to running, parse args, get command name and command args
func (app *App) prepareRun() (code int, name string) {
	// find command name.
	name = app.findCommandName(app.args)
	// is help command name.
	if name == HelpCommand {
		if len(app.args) == 0 { // like 'help'
			app.showApplicationHelp()
			return
		}

		var cmds []string
		if isValidCmdName(app.args[0]) {
			cmds = []string{app.args[0]}
		}

		// like 'help COMMAND'
		code = app.showCommandHelp(cmds)
		return
	}

	// not input and not set defaultCommand
	if name == "" {
		// run app.Func
		if app.Func != nil {
			code = app.doRunFunc(app.args)
			return
		}

		app.showApplicationHelp()
		return
	}

	// is not exist name.
	if app.inputName == "" {
		Logf(VerbDebug, "input the command is not an registered: %s", name)
		// display unknown input command and similar commands tips
		app.showCommandTips(name)
		return
	}

	// is valid command name.
	app.commandName = name
	return GOON, name
}

func (app *App) findCommandName(args []string) (name string) {
	// not input command, will try run app.defaultCommand
	if len(args) == 0 {
		name = app.defaultCommand

		// It is empty
		if name == "" {
			return
		}

		// It is not an valid command name.
		if false == app.IsCommand(name) {
			Logf(VerbError, "the default command '<cyan>%s</>' is invalid", name)
			return "" // invalid, return empty string.
		}

		return name
	}

	name = args[0]

	// check first arg is valid name string.
	if isValidCmdName(name) {
		realName := app.ResolveAlias(name)

		// is valid command name.
		if app.IsCommand(realName) {
			app.args = args[1:] // update args.
			app.inputName = name
			Debugf("input command: '<cyan>%s</>', real command: '<mga>%s</>'", name, realName)
		}

		return realName
	}

	return ""
}

// Run running application
//
// Usage:
//	// run with os.Args
//	app.Run(nil)
//	app.Run(os.Args[1:])
//	// custom args
//	app.Run([]string{"cmd", ...})
func (app *App) Run(args []string) (code int) {
	// ensure application initialized
	app.initialize()

	// if not set input args
	if args == nil {
		args = os.Args[1:] // exclude first arg, it's binFile.
	}

	Debugf("will begin run cli application. args: %v", args)

	// parse global flags
	if false == app.parseGlobalOpts(args) {
		return app.exitOnEnd(code)
	}

	Logf(VerbCrazy, "begin run console application, PID: %d", app.PID())

	var name string
	code, name = app.prepareRun()
	if code != GOON {
		return app.exitOnEnd(code)
	}

	// trigger event
	app.fireEvent(EvtAppPrepareAfter, app)

	// do run input command
	code = app.doRunCmd(name, app.args)

	Debugf("command '%s' run complete, exit with code: %d", name, code)
	return app.exitOnEnd(code)
}

func (app *App) doRunCmd(name string, args []string) (code int) {
	cmd := app.Command(name)
	app.fireEvent(EvtAppBefore, cmd.Copy())

	Debugf("will run command '%s' with args: %v", name, args)

	// contains keywords "-h" OR "--help" on end
	// if cmd.hasHelpKeywords() {
	// 	Logf(VerbDebug, "contains help keywords in flags, render command help message")
	// 	cmd.ShowHelp()
	// 	return
	// }

	// parse command options
	// args, err = cmd.parseOptions(args)

	// do execute command
	// if err := cmd.innerExecute(args, true); err != nil {
	if err := cmd.innerDispatch(args); err != nil {
		code = ERR
		app.fireEvent(EvtAppError, err)
	} else {
		app.fireEvent(EvtAppAfter, nil)
	}
	return
}

func (app *App) doRunFunc(args []string) (code int) {
	// app bind args TODO
	// app.ParseArgs(args)

	// do execute command
	if err := app.Func(app, args); err != nil {
		code = ERR
		app.fireEvent(EvtAppError, err)
	} else {
		app.fireEvent(EvtAppAfter, nil)
	}

	return
}

func (app *App) exitOnEnd(code int) int {
	Debugf("application exit with code: %d", code)

	if app.ExitOnEnd {
		app.Exit(code)
	}
	return code
}

// Exec running other command in current command
func (app *App) Exec(name string, args []string) (err error) {
	if !app.IsCommand(name) {
		return fmt.Errorf("exec unknown command name '%s'", name)
	}

	cmd := app.commands[name]

	// parse flags and execute command
	return cmd.innerExecute(args, false)
}

// CommandName get current command name
func (app *App) CommandName() string {
	return app.commandName
}

/*************************************************************
 * display app help
 *************************************************************/

// AppHelpTemplate help template for app(all commands)
var AppHelpTemplate = `{{.Desc}} (Version: <info>{{.Version}}</>)
<comment>Usage:</>
  {$binName} [Global Options...] <info>{command}</> [--option ...] [argument ...]

<comment>Global Options:</>
{{.GOpts}}
<comment>Available Commands:</>{{range $module, $cs := .Cs}}{{if $module}}
<comment> {{ $module }}</>{{end}}{{ range $cs }}
  <info>{{.Name | paddingName }}</> {{.Desc}}{{if .Aliases}} (alias: <cyan>{{ join .Aliases ","}}</>){{end}}{{end}}{{end}}

  <info>{{ paddingName "help" }}</> Display help information

Use "<cyan>{$binName} {COMMAND} -h</>" for more information about a command
`

// display app version info
func (app *App) showVersionInfo() {
	Debugf("print application version info")

	color.Printf(
		"%s\n\nVersion: <cyan>%s</>\n",
		strutil.UpperFirst(app.Desc),
		app.Version,
	)

	if app.Logo.Text != "" {
		color.Printf("%s\n", color.WrapTag(app.Logo.Text, app.Logo.Style))
	}
}

// display unknown input command and similar commands tips
func (app *App) showCommandTips(name string) {
	Debugf("show similar command tips")

	color.Error.Tips(`unknown input command "<mga>%s</>"`, name)
	if ns := app.findSimilarCmd(name); len(ns) > 0 {
		color.Printf("\nMaybe you mean:\n  <green>%s</>\n", strings.Join(ns, ", "))
	}

	color.Printf("\nUse <cyan>%s --help</> to see available commands\n", app.binName)
}

// display app help and list all commands
func (app *App) showApplicationHelp() {
	Debugf("render application commands list")

	// cmdHelpTemplate = color.ReplaceTag(cmdHelpTemplate)
	// render help text template
	s := helper.RenderText(AppHelpTemplate, map[string]interface{}{
		"Cs":    app.moduleCommands,
		"GOpts": app.gFlags.String(),
		// app version
		"Version": app.Version,
		// always upper first char
		"Desc": strutil.UpperFirst(app.Desc),
	}, template.FuncMap{
		"paddingName": func(n string) string {
			return strutil.PadRight(n, " ", app.nameMaxWidth)
		},
	})

	// parse help vars and render color tags
	color.Print(app.ReplaceVars(s))
}

// showCommandHelp display help for an command
func (app *App) showCommandHelp(list []string) (code int) {
	binName := app.binName
	if len(list) != 1 {
		color.Error.Tips("Too many arguments given.\n\nUsage: %s help {COMMAND}", binName)
		return ERR
	}

	// get real name
	name := app.cmdAliases.ResolveAlias(list[0])
	if name == HelpCommand || name == "-h" {
		color.Println("Display help message for application or command.\n")
		color.Printf("Usage:\n <cyan>%s {COMMAND} --help</> OR <cyan>%s help {COMMAND}</>\n", binName, binName)
		return
	}

	cmd, exist := app.commands[name]
	if !exist {
		color.Error.Prompt("Unknown command name '%s'. Run '<cyan>%s -h</>' see all commands", name, binName)
		return ERR
	}

	// show help for the command.
	cmd.ShowHelp()
	return
}

// show bash/zsh completion
func (app *App) showAutoCompletion(_ []string) {
	// TODO ...
}

// findSimilarCmd find similar cmd by input string
func (app *App) findSimilarCmd(input string) []string {
	var ss []string
	// ins := strings.Split(input, "")
	// fmt.Print(input, ins)
	ln := len(input)

	names := app.CmdNameMap()
	names["help"] = 4 // add 'help' command

	// find from command names
	for name := range names {
		cln := len(name)
		if cln > ln && strings.Contains(name, input) {
			ss = append(ss, name)
		} else if ln > cln && strings.Contains(input, name) {
			// sns := strings.Split(str, "")
			ss = append(ss, name)
		}

		// max find 5 items
		if len(ss) == 5 {
			break
		}
	}

	// find from aliases
	for alias := range app.cmdAliases {
		// max find 5 items
		if len(ss) >= 5 {
			break
		}

		cln := len(alias)
		if cln > ln && strings.Contains(alias, input) {
			ss = append(ss, alias)
		} else if ln > cln && strings.Contains(input, alias) {
			ss = append(ss, alias)
		}
	}

	return ss
}
