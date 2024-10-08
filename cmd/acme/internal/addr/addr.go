package addr

import (
	"plramos.win/9fans/cmd/acme/internal/alog"
	"plramos.win/9fans/cmd/acme/internal/regx"
	"plramos.win/9fans/cmd/acme/internal/runes"
)

const (
	None = '\x00'
	Fore = '+'
	Back = '-'
)

const (
	Char = iota
	Line
)

// Advance starts at q0 and advances nl lines,
// being careful not to walk past the end of the text,
// and then nr chars, being careful not to walk past
// the end of the current line.
// It returns the final position.
func Advance(t runes.Text, q0 int, nl int, nr int) int {
	for nl > 0 && q0 < t.Len() {
		tmp1 := q0
		q0++
		if t.RuneAt(tmp1) == '\n' {
			nl--
		}
	}
	if nl > 0 {
		return q0
	}
	for nr > 0 && q0 < t.Len() && t.RuneAt(q0) != '\n' {
		q0++
		nr--
	}
	return q0
}

func number(showerr bool, t runes.Text, r runes.Range, line int, dir rune, size int, evalp *bool) runes.Range {
	q0 := r.Pos
	q1 := r.End
	if size == Char {
		if dir == Fore {
			line = r.End + line
		} else if dir == Back {
			if r.Pos == 0 && line > 0 {
				r.Pos = t.Len()
			}
			line = r.Pos - line
		}
		if line < 0 || line > t.Len() {
			goto Rescue
		}
		*evalp = true
		return runes.Rng(line, line)
	}
	switch dir {
	case None:
		q0 = 0
		q1 = 0
		goto Forward
	case Fore:
		if q1 > 0 {
			for q1 < t.Len() && t.RuneAt(q1-1) != '\n' {
				q1++
			}
		}
		q0 = q1
		goto Forward
	case Back:
		if q0 < t.Len() {
			for q0 > 0 && t.RuneAt(q0-1) != '\n' {
				q0--
			}
		}
		q1 = q0
		for line > 0 && q0 > 0 {
			if t.RuneAt(q0-1) == '\n' {
				line--
				if line >= 0 {
					q1 = q0
				}
			}
			q0--
		}
		/* :1-1 is :0 = #0, but :1-2 is an error */
		if line > 1 {
			goto Rescue
		}
		for q0 > 0 && t.RuneAt(q0-1) != '\n' {
			q0--
		}
	}
Return:
	*evalp = true
	return runes.Rng(q0, q1)

Forward:
	for line > 0 && q1 < t.Len() {
		tmp2 := q1
		q1++
		if t.RuneAt(tmp2) == '\n' || q1 == t.Len() {
			line--
			if line > 0 {
				q0 = q1
			}
		}
	}
	if line == 1 && q1 == t.Len() { // 6 goes to end of 5-line file
		goto Return
	}
	if line > 0 {
		goto Rescue
	}
	goto Return

Rescue:
	if showerr {
		alog.Printf("address out of range\n")
	}
	*evalp = false
	return r
}

func regexp(showerr bool, t runes.Text, lim runes.Range, r runes.Range, pat []rune, dir rune, foundp *bool) runes.Range {
	if pat[0] == '\x00' && regx.Null() {
		if showerr {
			alog.Printf("no previous regular expression\n")
		}
		*foundp = false
		return r
	}
	if pat[0] != 0 && !regx.Compile(pat) {
		*foundp = false
		return r
	}
	var found bool
	var sel regx.Ranges
	if dir == Back {
		found = regx.MatchBackward(t, r.Pos, &sel)
	} else {
		var q int
		if lim.Pos < 0 {
			q = runes.Infinity
		} else {
			q = lim.End
		}
		found = regx.Match(t, nil, r.End, q, &sel)
	}
	if !found && showerr {
		alog.Printf("no match for regexp\n")
	}
	*foundp = found
	return sel.R[0]
}

func Eval(showerr bool, t runes.Text, lim runes.Range, ar runes.Range, a interface{}, q0 int, q1 int, getc func(interface{}, int) rune, evalp *bool, qp *int, reverse bool) runes.Range {
	r := ar
	q := q0
	dir := None
	size := Line
	var c rune
	for q < q1 {
		prevc := c
		c = getc(a, q)
		q++
		var nr runes.Range
		var pat []rune
		var n int
		var nc rune
		switch c {
		default:
			*qp = q - 1
			return r
		case ';':
			ar = r
			fallthrough
		/* fall through */
		case ',':
			if prevc == 0 { /* lhs defaults to 0 */
				r.Pos = 0
			}
			if q >= q1 && t != nil /* && t.file != nil */ { /* rhs defaults to $ */
				r.End = t.Len()
			} else {
				nr = Eval(showerr, t, lim, ar, a, q, q1, getc, evalp, &q, false)
				r.End = nr.End
			}
			*qp = q
			return r
		case '+', '-':
			if *evalp && (prevc == '+' || prevc == '-') {
				nc = getc(a, q)
				if nc != '#' && nc != '/' && nc != '?' {
					r = number(showerr, t, r, 1, prevc, Line, evalp) /* do previous one */
				}
			}
			dir = c
		case '.', '$':
			if q != q0+1 {
				*qp = q - 1
				return r
			}
			if *evalp {
				if c == '.' {
					r = ar
				} else {
					r = runes.Rng(t.Len(), t.Len())
				}
			}
			if q < q1 {
				dir = Fore
			} else {
				dir = None
			}
		case '#':
			if q == q1 || func() bool { c = getc(a, q); _r := c < '0'; q++; return _r }() || '9' < c {
				*qp = q - 1
				return r
			}
			size = Char
			fallthrough
		/* fall through */
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			n = int(c - '0')
			for q < q1 {
				nc = getc(a, q)
				q++
				if nc < '0' || '9' < nc {
					q--
					break
				}
				n = n*10 + int(nc-'0')
			}
			if *evalp {
				r = number(showerr, t, r, n, dir, size, evalp)
			}
			dir = None
			size = Line
		case '?':
			dir = Back
			fallthrough
		/* fall through */
		case '/':
			pat = nil
			for q < q1 {
				c = getc(a, q)
				q++
				switch c {
				case '\n':
					q--
					goto out
				case '\\':
					pat = append(pat, c)
					if q == q1 {
						goto out
					}
					c = getc(a, q)
					q++
				case '/':
					goto out
				}
				pat = append(pat, c)
			}
		out:
			if *evalp {
				r = regexp(showerr, t, lim, r, pat, dir, evalp)
			}
			dir = None
			if reverse {
				dir = Back
			}
			size = Line
		}
	}
	if *evalp && dir != None {
		r = number(showerr, t, r, 1, dir, Line, evalp) /* do previous one */
	}
	*qp = q
	return r
}
