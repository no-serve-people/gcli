package gcli

import (
	"context"
	"fmt"
	"strings"

	"github.com/gookit/gcli/v3/events"
	"github.com/gookit/goutil/errorx"
	"github.com/gookit/goutil/maputil"
)

// constants for hooks event, there are default allowed event names
const (
	EvtAppInit = events.OnAppInitAfter

	EvtAppPrepareAfter = events.OnAppPrepared

	EvtAppRunBefore = events.OnAppRunBefore
	EvtAppRunAfter  = events.OnAppRunAfter
	EvtAppRunError  = events.OnAppRunError

	EvtCmdInit = events.OnCmdInitAfter

	// EvtCmdNotFound app or sub command not found
	EvtCmdNotFound = events.OnCmdNotFound
	// EvtAppCmdNotFound app command not found
	EvtAppCmdNotFound = events.OnAppCmdNotFound
	// EvtCmdSubNotFound sub command not found
	EvtCmdSubNotFound = events.OnCmdSubNotFound

	EvtCmdOptParsed = events.OnCmdOptParsed

	// EvtCmdRunBefore cmd run
	EvtCmdRunBefore = events.OnCmdRunBefore
	EvtCmdRunAfter  = events.OnCmdRunAfter
	EvtCmdRunError  = events.OnCmdRunError

	// EvtCmdExecBefore cmd exec
	EvtCmdExecBefore = events.OnCmdExecBefore
	EvtCmdExecAfter  = events.OnCmdExecAfter
	EvtCmdExecError  = events.OnCmdExecError

	EvtGOptionsParsed = events.OnGlobalOptsParsed
)

// runErr struct
type runErr struct {
	code int
	err  error
}

// newRunErr instance
func newRunErr(code int, err error) errorx.ErrorCoder {
	return &runErr{code: code, err: err}
}

// Code for error
func (e *runErr) Code() int {
	return e.code
}

// Error string
func (e *runErr) Error() string {
	return fmt.Sprintf("%v with code %d", e.err, e.code)
}

// HookFunc definition.
//
// Returns:
//   - True  for stop continue run.
//   - False continue handle next logic.
type HookFunc func(ctx *HookCtx) (stop bool)

/*************************************************************
 * simple events manage
 *************************************************************/

// Hooks struct. hookManager
type Hooks struct {
	// Hooks can set some hooks func on running.
	hooks map[string]HookFunc
}

// On register event hook by name
func (h *Hooks) On(name string, handler HookFunc) {
	if handler == nil {
		panicf("event %q handler is nil", name)
	}

	if h.hooks == nil {
		h.hooks = make(map[string]HookFunc)
	}
	h.hooks[name] = handler
}

// AddHook register on not exists hook.
func (h *Hooks) AddHook(name string, handler HookFunc) {
	if _, ok := h.hooks[name]; !ok {
		h.On(name, handler)
	}
}

// Fire event by name, allow with event data.
// returns True for stop continue run.
func (h *Hooks) Fire(event string, ctx *HookCtx) (stop bool) {
	if handler, ok := h.hooks[event]; ok {
		return handler(ctx)
	}
	return false
}

// HasHook registered check.
func (h *Hooks) HasHook(event string) bool {
	_, ok := h.hooks[event]
	return ok
}

// ResetHooks clear all hooks
func (h *Hooks) ResetHooks() {
	h.hooks = nil
}

/*************************************************************
 * events context
 *************************************************************/

// HookCtx struct
type HookCtx struct {
	context.Context
	maputil.Data
	App *App
	Cmd *Command

	stop bool // stop continue handle.
	err  error
	name string
}

func newHookCtx(name string, c *Command, data map[string]any) *HookCtx {
	if data == nil {
		data = make(maputil.Data)
	}

	hc := &HookCtx{
		name: name,
		Cmd:  c,
		Data: data,
		// with empty context
		Context: context.Background(),
	}

	if c != nil {
		hc.App = c.app
	}
	return hc
}

// Err of event
func (hc *HookCtx) Err() error {
	return hc.err
}

// Name of event
func (hc *HookCtx) Name() string {
	return hc.name
}

// Stopped value
func (hc *HookCtx) Stopped() bool {
	return hc.stop
}

// SetStop value
func (hc *HookCtx) SetStop(stop bool) bool {
	hc.stop = stop
	return hc.stop
}

// WithErr value
func (hc *HookCtx) WithErr(err error) *HookCtx {
	hc.err = err
	return hc
}

// WithData to ctx
func (hc *HookCtx) WithData(data map[string]any) *HookCtx {
	if data != nil {
		hc.Data = data
	}
	return hc
}

// WithApp to ctx
func (hc *HookCtx) WithApp(a *App) *HookCtx {
	hc.App = a
	return hc
}

/*************************************************************
 * app/cmd help string-var replacer
 *************************************************************/

// HelpVarFormat allow string replace on render help info.
//
// Default support:
//
//	"{$binName}" "{$cmd}" "{$fullCmd}" "{$workDir}"
const HelpVarFormat = "{$%s}"

// HelpReplacer provide string var replace for render help template.
type HelpReplacer struct {
	VarOpen, VarClose string

	// replaces you can add string-var map for render help info.
	replaces map[string]string
}

// AddReplace get command name. AddReplace
func (hv *HelpReplacer) AddReplace(name, value string) {
	if hv.replaces == nil {
		hv.replaces = make(map[string]string)
	}
	hv.replaces[name] = value
}

// AddReplaces add multi tpl vars.
func (hv *HelpReplacer) AddReplaces(vars map[string]string) {
	for n, v := range vars {
		hv.AddReplace(n, v)
	}
}

// GetReplace get a help var by name
func (hv *HelpReplacer) GetReplace(name string) string {
	return hv.replaces[name]
}

// Replaces get all tpl vars.
func (hv *HelpReplacer) Replaces() map[string]string {
	return hv.replaces
}

// ReplacePairs replace string vars in the input text.
func (hv *HelpReplacer) ReplacePairs(input string) string {
	// if not use var
	if !strings.Contains(input, "{$") {
		return input
	}

	var ss []string
	for n, v := range hv.replaces {
		ss = append(ss, fmt.Sprintf(HelpVarFormat, n), v)
	}

	return strings.NewReplacer(ss...).Replace(input)
}
