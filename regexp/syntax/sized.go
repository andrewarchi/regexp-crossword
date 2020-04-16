package syntax

// sizedRegexp is a regular expression with bounded length.
type sizedRegexp struct {
	sizes    []*Regexp // len: max - min
	min, max int       // min <= max
}

func newSizedRegexp(min, max int) *sizedRegexp {
	return &sizedRegexp{make([]*Regexp, max-min), min, max}
}

func (s *sizedRegexp) inBounds(size int) bool {
	return s.min <= size && size < s.max
}

func (s *sizedRegexp) size(n int) *Regexp {
	if !s.inBounds(n) {
		return &Regexp{Op: OpNoMatch}
	}
	return s.sizes[n-s.min]
}

func (s *sizedRegexp) regexp() *Regexp {
	if s.min+1 == s.max {
		return s.sizes[0]
	}
	return &Regexp{Op: OpAlternate, Sub: s.sizes}
}

func (s *sizedRegexp) trim(min, max int) *sizedRegexp {
	if max < min || max < s.min || min > s.max {
		panic("regexp: invalid trim bounds")
	}
	t := &sizedRegexp{s.sizes, s.min, s.max}
	if min > t.min {
		t.sizes = t.sizes[min-t.min:]
		t.min = min
	}
	if max < t.max {
		t.sizes = t.sizes[:t.max-max]
		t.max = max
	}
	return t
}

func (s *sizedRegexp) insert(re *Regexp, size int) {
	if !s.inBounds(size) {
		panic("regexp: invalid insert size")
	}
	i := size - s.min
	if s.sizes[i] == nil {
		s.sizes[i] = re
	} else {
		if s.sizes[i].Op != OpAlternate {
			nre := &Regexp{Op: OpAlternate}
			nre.Sub = append(re.Sub0[:0], s.sizes[i])
			s.sizes[i] = nre
		}
		if re.Op == OpAlternate {
			s.sizes[i].Sub = append(s.sizes[i].Sub, re.Sub...)
		} else {
			s.sizes[i].Sub = append(s.sizes[i].Sub, re)
		}
	}
}

func concat(a, b *sizedRegexp, min, max int) *sizedRegexp {
	cMin, cMax := a.min+b.min, a.max+b.max
	if cMin >= max {
		return &sizedRegexp{nil, 0, 0}
	}
	if max < cMax {
		cMax = max
	}
	c := newSizedRegexp(cMin, cMax)
	for i, aRe := range a.sizes {
		if aRe != nil {
			for j, bRe := range b.sizes {
				if bRe != nil {
					n := i + j + cMin
					if n < cMax {
						if aRe.Op == OpEmptyMatch || bRe.Op == OpNoMatch {
							c.insert(bRe, n)
						} else if bRe.Op == OpEmptyMatch || aRe.Op == OpNoMatch {
							c.insert(aRe, n)
						} else if aRe.Op == OpLiteral && bRe.Op == OpLiteral {
							ab := &Regexp{Op: OpLiteral}
							ab.Rune = append(ab.Rune, aRe.Rune...)
							ab.Rune = append(ab.Rune, bRe.Rune...)
							c.insert(ab, n)
						} else {
							ab := &Regexp{Op: OpConcat}
							ab.Sub = ab.Sub0[:0]

							if aRe.Op == OpConcat {
								ab.Sub = append(ab.Sub, aRe.Sub...)
							} else {
								ab.Sub = append(ab.Sub, aRe)
							}
							if bRe.Op == OpConcat {
								ab.Sub = append(ab.Sub, bRe.Sub...)
							} else {
								ab.Sub = append(ab.Sub, bRe)
							}

							c.insert(ab, n)
						}
					}
				}
			}
		}
	}
	return c.trim(min, max)
}

func union(a, b *sizedRegexp, min, max int) *sizedRegexp {
	cMin, cMax := a.min, a.max
	if b.min < a.min {
		cMin = b.min
	}
	if b.max > a.max {
		cMax = b.max
	}
	if min > cMin {
		cMin = min
	}
	if max < cMax {
		cMax = max
	}

	c := newSizedRegexp(cMin, cMax)
	for i, aRe := range a.sizes {
		if aRe != nil && c.inBounds(i+a.min) {
			c.insert(aRe, i+a.min)
		}
	}
	for i, bRe := range b.sizes {
		if bRe != nil && c.inBounds(i+b.min) {
			c.insert(bRe, i+b.min)
		}
	}
	return c
}

func plus(a *sizedRegexp, min, max int) *sizedRegexp {
	aMin := a.min
	if a.min == 0 { // cannot divide by zero
		for i := 1; i < len(a.sizes); i++ {
			if a.sizes[i] != nil {
				aMin = a.min + i
				goto Nonzero
			}
		}
		return a.trim(min, max)
	}
Nonzero:

	// TODO: concat using shared terms like exponentiation by squaring.
	star := a
	acc := a
	n := (max + aMin - 1) / aMin
	for i := 2; i < n; i++ {
		acc = concat(acc, a, min, max)
		star = union(star, acc, min, max)
	}
	return star.trim(min, max)
}

func (re *Regexp) Reverse() *Regexp {
	if re == nil {
		return nil
	}
	switch re.Op {
	case OpNoMatch, OpEmptyMatch,
		OpBeginLine, OpEndLine, OpBeginText, OpEndText,
		OpWordBoundary, OpNoWordBoundary,
		OpCharClass, OpAnyCharNotNL, OpAnyChar:
		return re
	case OpLiteral:
		nre := &Regexp{Op: OpLiteral, Rune: make([]rune, len(re.Rune))}
		for i, r := range re.Rune {
			nre.Rune[len(re.Rune)-i-1] = r
		}
		return nre
	case OpBackref:
		return re
	case OpCapture, OpStar, OpPlus, OpQuest, OpRepeat:
		sub := re.Sub[0].Reverse()
		if sub == re.Sub[0] {
			return re
		}
		nre := new(Regexp)
		*nre = *re
		nre.Sub = append(nre.Sub0[:0], sub)
		return nre
	case OpConcat:
		if len(re.Sub) == 1 {
			return re.Sub[0].Reverse()
		}
		nre := &Regexp{Op: OpConcat, Sub: make([]*Regexp, len(re.Sub))}
		for i, sub := range re.Sub {
			nre.Sub[len(re.Sub)-i-1] = sub.Reverse()
		}
		return nre
	case OpAlternate:
		if len(re.Sub) == 1 {
			return re.Sub[0].Reverse()
		}
		// Simplify children, building new Regexp if children change.
		nre := re
		for i, sub := range re.Sub {
			nsub := sub.Reverse()
			if nre == re && nsub != sub {
				// Start a copy.
				nre = new(Regexp)
				*nre = *re
				nre.Rune = nil
				nre.Sub = append(nre.Sub0[:0], re.Sub[:i]...)
			}
			if nre != re {
				nre.Sub = append(nre.Sub, nsub)
			}
		}
		return nre
	default:
		panic("regexp: unhandled case in reverse")
	}
}

type constrainer struct {
	s        map[*Regexp]*sizedRegexp
	captures []*sizedRegexp
}

// on interval [min, max)
// min <= retmin < retmax <= max
func (re *Regexp) constrainLength(min, max int) *sizedRegexp {
	c := constrainer{make(map[*Regexp]*sizedRegexp), nil}
	return c.constrain(re, min, max)
}

func (c *constrainer) constrain(re *Regexp, min, max int) *sizedRegexp {
	if s, ok := c.s[re]; ok {
		return s.trim(min, max)
	}
	var s *sizedRegexp

	switch re.Op {
	case OpNoMatch, // TODO: should OpNoMatch have a special case?
		OpEmptyMatch,
		OpBeginLine, OpEndLine, OpBeginText, OpEndText,
		OpWordBoundary, OpNoWordBoundary:
		s = &sizedRegexp{[]*Regexp{re}, 0, 1}
	case OpCharClass, OpAnyCharNotNL, OpAnyChar:
		if min > 1 || max <= 1 {
			s = &sizedRegexp{nil, 0, 0}
			break
		}
		s = &sizedRegexp{[]*Regexp{re}, 1, 2}
	case OpLiteral:
		if min > len(re.Rune) || max <= len(re.Rune) {
			s = &sizedRegexp{nil, 0, 0}
			break
		}
		s = &sizedRegexp{[]*Regexp{re}, len(re.Rune), len(re.Rune) + 1}
	case OpCapture:
		capture := c.constrain(re.Sub[0], min, max)
		c.captures = append(c.captures, capture)
		s = capture
	case OpBackref:
		if re.Cap > len(c.captures) || re.Cap <= 0 {
			panic("regexp: capture not found")
		}
		s = c.captures[re.Cap-1]
	case OpStar:
		sub := c.constrain(re.Sub[0], 0, max)
		acc := plus(sub, min, max)
		if min == 0 {
			empty := &sizedRegexp{[]*Regexp{&Regexp{Op: OpEmptyMatch}}, 0, 1}
			acc = union(acc, empty, min, max)
		}
		s = acc.trim(min, max)
	case OpPlus:
		sub := c.constrain(re.Sub[0], 0, max)
		acc := plus(sub, min, max)
		s = acc.trim(min, max)
	case OpQuest:
		sub := c.constrain(re.Sub[0], min, max)
		if min == 0 {
			empty := &sizedRegexp{[]*Regexp{&Regexp{Op: OpEmptyMatch}}, 0, 1}
			sub = union(sub, empty, min, max)
		}
		s = sub
	case OpConcat:
		if len(re.Sub) == 0 {
			s = &sizedRegexp{[]*Regexp{&Regexp{Op: OpEmptyMatch}}, 0, 1}
			break
		}
		acc := c.constrain(re.Sub[0], 0, max)
		for _, sub := range re.Sub[1:] {
			acc = concat(acc, c.constrain(sub, 0, max-acc.min), min, max)
		}
		s = acc.trim(min, max)
	case OpAlternate:
		if len(re.Sub) == 0 {
			s = &sizedRegexp{nil, 0, 0}
			break
		}
		acc := c.constrain(re.Sub[0], min, max)
		for _, sub := range re.Sub[1:] {
			acc = union(acc, c.constrain(sub, min, max), min, max)
		}
		s = acc
	case OpRepeat:
		panic("regexp: repeat not simplified")
	default:
		panic("regexp: unhandled case in constrain")
	}

	c.s[re] = s
	return s
}
