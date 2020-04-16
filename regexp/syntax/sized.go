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
		panic("concat overflow")
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
				goto nonzero
			}
		}
		return a.trim(min, max)
	}
nonzero:

	// TODO: concat using shared terms like exponentiation by squaring.
	acc := a
	n := (max + aMin - 1) / aMin
	for i := 1; i < n; i++ {
		acc = union(acc, concat(acc, a, min, max), min, max)
	}
	return acc.trim(min, max)
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
		if re.Cap > len(c.captures) {
			panic("regexp: capture not found")
		}
		s = c.captures[re.Cap].trim(min, max)
	case OpStar:
		acc := plus(c.constrain(re.Sub[0], min, max), min, max)
		if acc.inBounds(0) {
			acc.insert(&Regexp{Op: OpEmptyMatch}, 0)
		}
		s = acc
	case OpPlus:
		s = plus(c.constrain(re.Sub[0], min, max), min, max)
	case OpQuest:
		sub := c.constrain(re.Sub[0], min, max)
		if sub.inBounds(0) {
			sub.insert(&Regexp{Op: OpEmptyMatch}, 0)
		}
		s = sub
	case OpConcat:
		if len(re.Sub) == 0 {
			s = &sizedRegexp{[]*Regexp{&Regexp{Op: OpEmptyMatch}}, 0, 1}
			break
		}
		acc := c.constrain(re.Sub[0], min, max)
		for _, sub := range re.Sub[1:] {
			// TODO
			// acc = concat(acc, c.constrain(sub, min-acc.min, max-acc.min), min, max)
			acc = concat(acc, c.constrain(sub, min, max), min, max)
		}
		s = acc
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
