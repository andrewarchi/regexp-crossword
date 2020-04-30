package crossword

// MIT Puzzle
// https://www.mit.edu/~puzzle/2013/coinheist.com/rubik/a_regular_crossword/grid.pdf
var mitPuzzle = Puzzle{
	ID:   "mit",
	Size: 7,
	PatternsX: [][]string{{
		`(ND|ET|IN)[^X]*`,
		`[CHMNOR]*I[CHMNOR]*`,
		`P+(..)\1.*`,
		`(E|CR|MN)*`,
		`([^MC]|MM|CC)*`,
		`[AM]*CM(RC)*R?`,
		`.*`,
		`.*PRR.*DDC.*`,
		`(HHX|[^HX])*`,
		`([^EMC]|EM)*`,
		`.*OXR.*`,
		`.*LR.*RL.*`,
		`.*SE.*UE.*`,
	}},
	PatternsY: [][]string{{
		`.*H.*H.*`,
		`(DI|NS|TH|OM)*`,
		`F.*[AO].*[AO].*`,
		`(O|RHH|MM)*`,
		`.*`,
		`C*MC(CCC|MM)*`,
		`[^C]*[^R]*III.*`,
		`(...?)\1*`,
		`([^X]|XCC)*`,
		`(RR|HHH)*.?`,
		`N.*X.X.X.*E`,
		`R*D*M*`,
		`.(C|HH)*`,
	}},
	PatternsZ: [][]string{{
		`.*G.*V.*H.*`,
		`[CR]*`,
		`.*XEXM*`,
		`.*DD.*CCM.*`,
		`.*XHCR.*X.*`,
		`.*(.)(.)(.)(.)\4\3\2\1.*`,
		`.*(IN|SE|HI)`,
		`[^C]*MMM[^C]*`,
		`.*(.)C\1X\1.*`,
		`[CEIMU]*OH[AEMOR]*`,
		`(RX|[^R])*`,
		`[^M]*M[^M]*`,
		`(S|MM|HHH)*`,
	}},
	Hexagonal: true,
}