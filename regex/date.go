package regex

import (
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ylacancellera/galera-log-explainer/utils"
)

// 5.5 date : 151027  6:02:49
// 5.6 date : 2019-07-17 07:16:37
//5.7 date : 2019-07-17T15:16:37.123456Z
//5.7 date : 2019-07-17T15:16:37.123456+01:00
// 10.3 date: 2019-07-15  7:32:25
var DateLayouts = []string{
	"2006-01-02T15:04:05.000000Z",      // 5.7
	"2006-01-02T15:04:05.000000-07:00", // 5.7
	"060102 15:04:05",                  // 5.5
	"2006-01-02 15:04:05",              // 5.6
	"2006-01-02  15:04:05",             // 10.3
}

// BetweenDateRegex generate a regex to filter mysql error log dates to just get
// events between 2 dates
// Currently limited to filter by day to produce "short" regexes. Finer events will be filtered later in code
// Trying to filter hours, minutes using regexes would produce regexes even harder to read
// while not really adding huge benefit as we do not expect so many events of interets
func BetweenDateRegex(since, until *time.Time) string {
	/*
		"2006-01-02
		"2006-01-0[3-9]
		"2006-01-[1-9][0-9]
		"2006-0[2-9]-[0-9]{2}
		"2006-[1-9][0-9]-[0-9]{2}
		"200[7-9]-[0-9]{2}-[0-9]{2}
		"20[1-9][0-9]-[0-9]{2}-[0-9]{2}
	*/
	regexConstructor := []struct {
		unit      int
		unitToStr string
	}{
		{
			unit:      since.Day(),
			unitToStr: fmt.Sprintf("%02d", since.Day()),
		},
		{
			unit:      int(since.Month()),
			unitToStr: fmt.Sprintf("%02d", since.Month()),
		},
		{
			unit:      since.Year(),
			unitToStr: fmt.Sprintf("%d", since.Year())[2:],
		},
	}
	s := ""
	for _, layout := range []string{"2006-01-02", "060102"} {
		// base complete date
		lastTransformed := since.Format(layout)
		s += "|^" + lastTransformed

		for _, construct := range regexConstructor {
			if construct.unit != 9 {
				s += "|^" + utils.StringsReplaceReversed(lastTransformed, construct.unitToStr, string(construct.unitToStr[0])+"["+strconv.Itoa(construct.unit%10+1)+"-9]", 1)
			}
			// %1000 here is to cover the transformation of 2022 => 22
			s += "|^" + utils.StringsReplaceReversed(lastTransformed, construct.unitToStr, "["+strconv.Itoa((construct.unit%1000/10)+1)+"-9][0-9]", 1)

			lastTransformed = utils.StringsReplaceReversed(lastTransformed, construct.unitToStr, "[0-9][0-9]", 1)

		}
	}
	s += ")"
	return "(" + s[1:]
}

func NoDatesRegex() string {
	//return "((?![0-9]{4}-[0-9]{2}-[0-9]{2})|(?![0-9]{6}))"
	return "^(?![0-9]{4})"
}

/*
SYSLOG_DATE="\(Jan\|Feb\|Mar\|Apr\|May\|Jun\|Jul\|Aug\|Sep\|Oct\|Nov\|Dec\) \( \|[0-9]\)[0-9] [0-9]\{2\}:[0-9]\{2\}:[0-9]\{2\}"
REGEX_LOG_PREFIX="$REGEX_DATE \?[0-9]* "
*/

const k8sprefix = `{"log":"`

func SearchDateFromLog(logline string) (time.Time, string, bool) {
	if logline[:len(k8sprefix)] == k8sprefix {
		logline = logline[len(k8sprefix):]
	}
	for _, layout := range DateLayouts {
		if len(logline) < len(layout) {
			continue
		}
		t, err := time.Parse(layout, logline[:len(layout)])
		if err == nil {
			return t, layout, true
		}
	}
	log.Debug().Str("log", logline).Msg("could not find date from log")
	return time.Time{}, "", false
}
