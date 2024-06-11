package hw03frequencyanalysis

import (
	"cmp"
	"regexp"
	"slices"
	"strings"
)

// Top10 returns up to 10 words from string s sorted by using frequency + lexicographic order.
func Top10(s string) []string {
	// words holds all words from the input string
	words := strings.Fields(s)

	// wordFreq (word:freq) holds frequency for each unique word.
	wordFreq := map[string]int{}

	// Calculate frequency for each word
	for _, w := range words {
		w = cleanWord(strings.ToLower(w))
		if w != "" {
			wordFreq[w]++
		}
	}

	// Store unique words
	uniqWords := make([]string, 0, len(wordFreq))
	for w := range wordFreq {
		uniqWords = append(uniqWords, w)
	}

	slices.SortFunc(uniqWords, func(a string, b string) int {
		// 1. sort by frequency
		if n := cmp.Compare(wordFreq[b], wordFreq[a]); n != 0 {
			return n
		}

		// 2. sort in lexicographic order
		return cmp.Compare(a, b)
	})

	// Return up to 10 words
	return uniqWords[:min(10, len(uniqWords))]
}

var (
	punctMarkRe  = regexp.MustCompile("^[,.?!]|[,.?!]$")
	quote1MarkRe = regexp.MustCompile("^'(.+)'$")
	quote2MarkRe = regexp.MustCompile("^`(.+)`$")
	quote3MarkRe = regexp.MustCompile("^\"(.+)\"$")
)

// cleanWord removes punctuation marks at the start and the end of the word.
// Returns clear word. Special case: "-" is not a word, empty string will be returned.
func cleanWord(w string) string {
	if w == "" || w == "-" {
		return ""
	}

	wb := []byte(w)

	for _, cond := range []struct {
		re      *regexp.Regexp
		replace []byte
	}{
		{punctMarkRe, []byte{}},
		{quote1MarkRe, []byte("$1")},
		{quote2MarkRe, []byte("$1")},
		{quote3MarkRe, []byte("$1")},
	} {
		if cond.re.Match(wb) {
			wb = cond.re.ReplaceAll(wb, cond.replace)
			return string(wb)
		}
	}

	return w
}
