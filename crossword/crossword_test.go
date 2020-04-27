package crossword

import "testing"

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
