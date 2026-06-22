package llm

import (
	"strings"
)

// EstimateTokens calculates an approximate token count for a given text.
// Uses a standard heuristic: 1 token is roughly 4 characters or 0.75 words.
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}

	charCount := len(text)
	wordCount := len(strings.Fields(text))

	// Estimate based on characters (1 token per 4 chars)
	estFromChars := charCount / 4

	// Estimate based on words (4 tokens per 3 words)
	estFromWords := (wordCount * 4) / 3

	// Return the maximum of both estimates to be conservative, at least 1
	res := estFromChars
	if estFromWords > res {
		res = estFromWords
	}
	if res < 1 {
		return 1
	}
	return res
}
