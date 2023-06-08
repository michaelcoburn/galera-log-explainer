package main

import (
	"github.com/pkg/errors"
	"github.com/ylacancellera/galera-log-explainer/display"
	"github.com/ylacancellera/galera-log-explainer/regex"
	"github.com/ylacancellera/galera-log-explainer/types"
)

type list struct {
	// Paths is duplicated because it could not work as variadic with kong cli if I set it as CLI object
	Paths                  []string `arg:"" name:"paths" help:"paths of the log to use"`
	SkipStateColoredColumn bool     `help:"avoid having the placeholder colored with mysql state, which is guessed using several regexes that will not be displayed"`
	All                    bool     `help:"List everything" xor:"states,views,events,sst"`
	States                 bool     `help:"List WSREP state changes(SYNCED, DONOR, ...)" xor:"states"`
	Views                  bool     `help:"List how Galera views evolved (who joined, who left)" xor:"views"`
	Events                 bool     `help:"List generic mysql events (start, shutdown, assertion failures)" xor:"events"`
	SST                    bool     `help:"List Galera synchronization event" xor:"sst"`
}

func (l *list) Help() string {
	return `List events for each nodes in a columnar output
	It will merge logs between themselves

	"identifier" is an internal metadata, this is used to merge logs.

Usage:
	galera-log-explainer list --all <list of files>
	galera-log-explainer list --all *.log
	galera-log-explainer list --sst --views --states <list of files>
	galera-log-explainer list --events --views *.log
	`
}

func (l *list) Run() error {

	if !(l.All || l.Events || l.States || l.SST || l.Views) {
		return errors.New("Please select a type of logs to search: --all, or any parameters from: --sst --views --events --states")
	}

	toCheck := l.regexesToUse()

	timeline, err := timelineFromPaths(CLI.List.Paths, toCheck, CLI.Since, CLI.Until)
	if err != nil {
		return errors.Wrap(err, "Could not list events")
	}

	display.TimelineCLI(timeline, CLI.Verbosity)

	return nil
}

func (l *list) regexesToUse() types.RegexMap {

	// IdentRegexes is always needed: we would not be able to identify the node where the file come from
	toCheck := regex.IdentsMap
	if l.States || l.All {
		toCheck.Merge(regex.StatesMap)
	} else if !l.SkipStateColoredColumn {
		regex.SetVerbosity(types.DebugMySQL, regex.StatesMap)
		toCheck.Merge(regex.StatesMap)
	}
	if l.Views || l.All {
		toCheck.Merge(regex.ViewsMap)
	}
	if l.SST || l.All {
		toCheck.Merge(regex.SSTMap)
	}
	if l.Events || l.All {
		toCheck.Merge(regex.EventsMap)
	} else if !l.SkipStateColoredColumn {
		regex.SetVerbosity(types.DebugMySQL, regex.EventsMap)
		toCheck.Merge(regex.EventsMap)
	}
	return toCheck
}
