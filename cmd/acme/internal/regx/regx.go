package regx

import (
	"io"
	"regexp"
	"slices"

	"plramos.win/9fans/cmd/acme/internal/runes"
)

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

// Compile the regex into regex global variable. Flags are the golang regexp flags
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

// MatchAll matches the entire chunk of test.
// Match reads the runes one by one for matching, Match has to load everything into a
// slice.
func MatchAll(t runes.Text, startp, eof int) (sels []Ranges, found bool) {
	if eof == runes.Infinity {
		eof = t.Len()
	}

	s := t.Slice(startp, eof)

	matches := regex.FindAllStringSubmatchIndex(string(s), -1)
	if len(matches) == 0 {
		sel := Ranges{R: make([]runes.Range, 1)}
		sel.R = make([]runes.Range, 1, 1)
		sel.R[0] = runes.Range{Pos: -1, End: -1}
		sels = append(sels, sel)
		return sels, false
	}
	sels = make([]Ranges, 0, len(matches))
	for _, m := range matches {
// 		if m[1] >= len(s) {
// 			continue
// 		}
		sel := Ranges{R: make([]runes.Range, len(m)/2)}
		for i := range sel.R {
			sel.R[i] = runes.Range{Pos: m[2*i] + startp, End: m[2*i+1] + startp}
		}
		sels = append(sels, sel)
	}
	return sels, len(sels) > 0
}

// Matches a regexp in provided rune slice
func Match(t runes.Text, startp, eof int) (sel Ranges, found bool) {
	wrap := false
	if eof == runes.Infinity {
		eof = t.Len()
		wrap = true
	}
	rr := runes.NewRuneReader(t, startp, eof)
	sel.R = make([]runes.Range, 1, 10)
	sel.R[0] = runes.Range{Pos: -1, End: -1}

	m := regex.FindReaderSubmatchIndex(rr)
	matched := len(m) > 0

	if !matched && wrap {
		return Match(t, 0, runes.Infinity)
	}
	if !matched {
		Sel = sel
		return sel, false
	}

	sel.R[0] = runes.Range{Pos: m[0] + rr.Start(), End: m[1] + rr.Start()}
	for i := 2; i < len(m); i += 2 {
		r := runes.Range{Pos: m[i] + rr.Start(), End: m[i+1] + rr.Start()}
		sel.R = append(sel.R, r)
	}
	Sel = sel

	return sel, true
}

// MatchLines divides t runes.Text into lines before matching.
// returns all the marches across all the lines
func MatchLines(t runes.Text, startp, eof int) (sels []Ranges, found bool) {
	rr := runes.NewRuneReader(t, startp, eof)

	eof = min(t.Len(), eof)
	if startp >= eof {
		return nil, false
	}

	for p := startp; p < eof; p++ {
		r, _, err := rr.ReadRune()
		if err == io.EOF {
			eof = p
			break
		}
		if r == '\n' {
			// we only want to match after the newline, unless there is no prevnl
			if sel, matched := MatchAll(t, startp, p); matched {
				sels = append(sels, sel...)
			}
			startp = p + 1
		}
	}
	if sel, matched := MatchAll(t, startp, eof); matched {
		sels = append(sels, sel...)
	}
	return sels, len(sels) > 0
}

// MatchBackward does a backwards regex match.
// Accept a rune slice contain all the text in which to match the regex and searchs from the end to the beginning of the slice.
func MatchBackward(t runes.Text, startp int) (sel Ranges, found bool) {
	defer func() { Sel = sel }()

	sel.R = make([]runes.Range, 0, 10)
	buf := make([]rune, t.Len())

	// matches from FindSubMarchIndex
	var matchedIn []int

	p := startp
	for p > 0 {
		buf[p] = t.RuneAt(p)
		matchedIn = regex.FindStringSubmatchIndex(string(buf[p:]))
		if len(matchedIn) != 0 { // found match
			break
		}
		p--
	}
	// no match
	if len(matchedIn) == 0 {
		sel.R = append(sel.R, runes.Range{Pos: -1, End: -1})
		return sel, false
	}

	for i := 0; i < len(matchedIn); i += 2 {
		r := runes.Range{Pos: matchedIn[i] + p, End: matchedIn[i+1] + p}
		sel.R = append(sel.R, r)
	}
	return sel, true
}
