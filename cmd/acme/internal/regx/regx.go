package regx

import (
	"regexp"
	"slices"

	"plramos.win/9fans/cmd/acme/internal/runes"
)

// TODO: Which is faster? Create a slice with all the runes have a
// slidding window or prepend/append to buffer. Match reads everything
// into a []rune and Matchs that, MatchBackward does a prepend&match
// buffer (buffer grows backwards). Match just allocs a buffer the size
// of the Text and fills it as needed.


var (
	// current compiled regex
	regex *regexp.Regexp

	// previous regexp compiled
	lastregex []rune
)

type Ranges struct{ R []runes.Range }

// Sel is the result of the Match or BackwardMatch
var Sel Ranges

// Null checks if there is any compiled regex
func Null() bool { return regex == nil }

func Compile(r []rune) bool {
	r = append([]rune("(?m)"), r...)
	if slices.Compare(r, lastregex) == 0 {
		return true
	}
	rgx, err := regexp.Compile(string(r))
	if err != nil {
		return false
	}
	regex = rgx
	lastregex = r
	Sel = Ranges{R: []runes.Range{{Pos: -1, End: -1}}}
	return true
}

// Matches a regexp in provided rune slice
func Match(t runes.Text, startp, eof int) (sel Ranges, found bool) {
	sel.R = make([]runes.Range, 1, 10)
	sel.R[0] = runes.Range{Pos: -1, End: -1}
	wrap := false
	if eof == runes.Infinity {
		eof = t.Len()
		wrap = true
	}
	r := t.Slice(startp, eof)
	m := regex.FindStringSubmatchIndex(string(r))
	if len(m) == 0 && wrap {
		// must restart from top, so I just create a []rune with everything from t
		// and call Match on that
		r = append(t.Slice(0, startp), r...)
		return Match(runes.Runes(r), 0, eof)
	} else if len(m) == 0 {
		Sel = sel
		return sel, false
	}

	sel.R[0] = runes.Range{Pos: m[0] + startp, End: m[1] + startp}
	for i := 2; i < len(m); i += 2 {
		r := runes.Range{Pos: m[i] + startp, End: m[i+1] + startp}
		sel.R = append(sel.R, r)
	}
	Sel = sel
	return sel, true
}

// MatchBackward does a backwards regex match.
// Accept a rune slice contain all the text in which to match the regex and searchs from the end to the beginning of the slice.
func MatchBackward(t runes.Text, startp int) (sel Ranges, found bool) {
	defer func() { Sel = sel }()

	sel.R = make([]runes.Range, 0, 10)
	buf := make([]rune, t.Len())

	// matches from FindSubMarchIndex
	var m []int

	p := startp
	for p > 0 {
		buf[p] = t.RuneAt(p)
		m = regex.FindStringSubmatchIndex(string(buf[p:]))
		if len(m) != 0 { // found match
			break
		}
		p--
	}
	// no match
	if len(m) == 0 {
		sel.R = append(sel.R, runes.Range{Pos: -1, End: -1})
		return sel, false
	}

	for i := 0; i < len(m); i += 2 {
		r := runes.Range{Pos: m[i] + p, End: m[i+1] + p}
		sel.R = append(sel.R, r)
	}
	return sel, true
}
