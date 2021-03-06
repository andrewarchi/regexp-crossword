package crossword

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andrewarchi/regexp-crossword/regexp/syntax"
)

// Challenge is a group of puzzles of similar difficulty.
type Challenge struct {
	ID            string    `json:"id"`
	Date          time.Time `json:"date"`
	Name          string    `json:"name"`
	Description   string    `json:"descr"`
	AchievementID string    `json:"achievement_id"`
	Puzzles       []Puzzle  `json:"puzzles"`
}

// Puzzle is a regular expression crossword puzzle.
type Puzzle struct {
	ID          string     `json:"id"`
	PlayerNo    int64      `json:"playerNo"`
	Name        string     `json:"name"`
	PatternsX   [][]string `json:"patternsX"`
	PatternsY   [][]string `json:"patternsY"`
	PatternsZ   [][]string `json:"patternsZ"`
	SolutionMap []int      `json:"solutionMap"`
	Characters  []string   `json:"characters"`
	Size        int        `json:"size"`
	Hexagonal   bool       `json:"hexagonal"`
	Mobile      bool       `json:"mobile"`
	Published   bool       `json:"published"`
	DateCreated UnixTime   `json:"dateCreated"`
	DateUpdated UnixTime   `json:"dateUpdated"`
	RatingAvg   float64    `json:"ratingAvg"`
	Votes       int64      `json:"votes"`
	Solved      UnixTime   `json:"solved"`
	Ambiguous   bool       `json:"ambiguous"`
}

// GetChallenges fetches all default challenges.
func GetChallenges() ([]Challenge, error) {
	res, err := http.Get("https://regexcrossword.com/data/challenges.json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var c []Challenge
	if err := json.NewDecoder(res.Body).Decode(&c); err != nil {
		return nil, err
	}
	return c, nil
}

// GetPlayerPuzzles fetches all user-submitted puzzles.
func GetPlayerPuzzles() ([]Puzzle, error) {
	res, err := http.Get("https://regexcrossword.com/api/puzzles")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var p []Puzzle
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, err
	}
	return p, nil
}

// SyntaxError is a pattern parse error.
type SyntaxError struct {
	Pattern string
	Err     error
}

// ValidatePatterns parses each pattern and reports syntax errors.
func (p *Puzzle) ValidatePatterns() []SyntaxError {
	var errs []SyntaxError
	for _, axis := range [3][][]string{p.PatternsX, p.PatternsY, p.PatternsZ} {
		for _, set := range axis {
			for _, pattern := range set {
				if _, err := syntax.Parse(pattern, syntax.Perl|syntax.Backref|syntax.PermissiveEscapes); err != nil {
					errs = append(errs, SyntaxError{pattern, err})
				}
			}
		}
	}
	return errs
}

func (p *Puzzle) PatternOps(counts map[syntax.Op]int) {
	for _, axis := range [3][][]string{p.PatternsX, p.PatternsY, p.PatternsZ} {
		for _, set := range axis {
			for _, pattern := range set {
				re, err := syntax.Parse(pattern, syntax.Perl|syntax.Backref|syntax.PermissiveEscapes)
				if err != nil {
					continue
				}
				countOps(re, counts)
			}
		}
	}
}

func countOps(re *syntax.Regexp, counts map[syntax.Op]int) {
	counts[re.Op]++
	for _, sub := range re.Sub {
		countOps(sub, counts)
	}
}
