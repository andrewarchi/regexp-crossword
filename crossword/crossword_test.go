package crossword

import (
	"fmt"
	"testing"

	"github.com/andrewarchi/regexp-crossword/regexp/syntax"
)

func TestGetChallenges(t *testing.T) {
	if _, err := GetChallenges(); err != nil {
		t.Fatal(err)
	}
}

func TestGetPlayerPuzzles(t *testing.T) {
	if _, err := GetPlayerPuzzles(); err != nil {
		t.Fatal(err)
	}
}

func TestValidatePatterns(t *testing.T) {
	challenges, err := GetChallenges()
	if err != nil {
		t.Fatal(err)
	}
	puzzles, err := GetPlayerPuzzles()
	if err != nil {
		t.Fatal(err)
	}
	var errs []SyntaxError
	for _, c := range challenges {
		for _, p := range c.Puzzles {
			errs = append(errs, p.ValidatePatterns()...)
		}
	}
	for _, p := range puzzles {
		errs = append(errs, p.ValidatePatterns()...)
	}
	if len(errs) != 0 {
		t.Fail()
	}
	grouped := make(map[string][]string)
	for _, err := range errs {
		e := err.Err.Error()
		grouped[e] = append(grouped[e], err.Pattern)
	}
	for err, patterns := range grouped {
		fmt.Println(err)
		for _, p := range patterns {
			fmt.Println(p)
		}
		fmt.Println()
	}
}

func TestOpUsage(t *testing.T) {
	challenges, err := GetChallenges()
	if err != nil {
		t.Fatal(err)
	}
	puzzles, err := GetPlayerPuzzles()
	if err != nil {
		t.Fatal(err)
	}
	counts := make(map[syntax.Op]int)
	for _, c := range challenges {
		for _, p := range c.Puzzles {
			p.PatternOps(counts)
		}
	}
	for _, p := range puzzles {
		p.PatternOps(counts)
	}
	t.Fail()
	fmt.Println(counts)
}
