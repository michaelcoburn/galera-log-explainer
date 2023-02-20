package main

import (
	"github.com/ylacancellera/galera-log-explainer/display"
	"github.com/ylacancellera/galera-log-explainer/regex"
	"github.com/ylacancellera/galera-log-explainer/types"
	"github.com/ylacancellera/galera-log-explainer/utils"
)

type list struct {
	// Paths is duplicated because it could not work as variadic with kong cli if I set it as CLI object
	Paths                  []string        `arg:"" name:"paths" help:"paths of the log to use"`
	Format                 string          `help:"Types of output format" enum:"cli,svg" default:"cli"`
	Verbosity              types.Verbosity `default:"1" help:"0: Info, 1: Detailed, 2: DebugMySQL (every mysql info the tool used), 3: Debug (internal tool debug)"`
	SkipStateColoredColumn bool            `help:"avoid having the placeholder colored with mysql state, which is guessed using several regexes that will not be displayed"`
	All                    bool            `help:"List everything" xor:"states,views,events,sst"`
	States                 bool            `help:"List WSREP state changes(SYNCED, DONOR, ...)" xor:"states"`
	Views                  bool            `help:"List how Galera views evolved (who joined, who left)" xor:"views"`
	Events                 bool            `help:"List generic mysql events (start, shutdown, assertion failures)" xor:"events"`
	SST                    bool            `help:"List Galera synchronization event" xor:"sst"`
}

func (l *list) Help() string {
	return "List events for each nodes"
}

func (l *list) Run() error {

	toCheck := l.regexesToUse()
	if CLI.List.Format == "svg" {
		// svg text does not handle cli special characters
		utils.SkipColor = true
	}

	timeline := timelineFromPaths(CLI.List.Paths, toCheck, CLI.Since, CLI.Until)

	switch CLI.List.Format {
	case "cli":
		display.TimelineCLI(timeline, CLI.List.Verbosity)
		break
	case "svg":
		display.TimelineSVG(timeline, CLI.List.Verbosity)
	}

	return nil
}

func (l *list) regexesToUse() []regex.LogRegex {

	// IdentRegexes is always needed: we would not be able to identify the node where the file come from
	toCheck := regex.IdentRegexes
	if CLI.List.States || CLI.List.All {
		toCheck = append(toCheck, regex.StatesRegexes...)
	} else if !CLI.List.SkipStateColoredColumn {
		toCheck = append(toCheck, regex.SetVerbosity(types.DebugMySQL, regex.StatesRegexes...)...)
	}
	if CLI.List.Views || CLI.List.All {
		toCheck = append(toCheck, regex.ViewsRegexes...)
	}
	if CLI.List.SST || CLI.List.All {
		toCheck = append(toCheck, regex.SSTRegexes...)
	}
	if CLI.List.Events || CLI.List.All {
		toCheck = append(toCheck, regex.EventsRegexes...)
	} else if !CLI.List.SkipStateColoredColumn {
		toCheck = append(toCheck, regex.SetVerbosity(types.DebugMySQL, regex.EventsRegexes...)...)
	}
	return toCheck
}
