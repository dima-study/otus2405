package hw03frequencyanalysis

import (
	"cmp"
	"slices"
	"strings"
)

// Top10 returns up to 10 words from string s sorted by using frequency + lexicographic order.
func Top10(s string) []string {
	// words holds all words from the input string
	words := strings.Fields(s)

	// wordFreq (word:freq) holds frequency for each unique word.
	wordFreq := map[string]int{}
	for _, w := range words {
		wordFreq[w]++
	}

	// Calculate frequency for each word
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
