package crossword

import (
	"fmt"
	"testing"
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
