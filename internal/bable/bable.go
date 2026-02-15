// Package bable is the Markov Babler algoritm
package bable

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
)

// Prefix is a Markov chain prefix of one or more words.
type Prefix []string

// String returns the Prefix as a string (for use as a map key).
func (p Prefix) String() string {
	return strings.Join(p, " ")
}

// Shift removes the first word from the Prefix and appends the given word.
func (p Prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

// Chain contains a map ("chain") of prefixes to a list of suffixes.
// A prefix is a string of prefixLen words joined with spaces.
// A suffix is a single word. A prefix can have multiple suffixes.
type Chain struct {
	chain     map[string][]string
	prefixLen int
}

// NewChain returns a new Chain with prefixes of prefixLen words.
func NewChain(prefixLen int) *Chain {
	return &Chain{make(map[string][]string), prefixLen}
}

// Build reads text from the provided Reader and
// parses it into prefixes and suffixes that are stored in Chain.
func (c *Chain) Build(r string) {
	br := strings.NewReader(r)
	p := make(Prefix, c.prefixLen)
	for {
		var s string
		if _, err := fmt.Fscan(br, &s); err != nil {
			break
		}
		key := p.String()
		c.chain[key] = append(c.chain[key], s)
		p.Shift(s)
	}
}

// Generate returns a string of at most n words generated from Chain.
func (c *Chain) Generate(n int) string {
	p := make(Prefix, c.prefixLen)
	var words []string
	for range n {
		choices := c.chain[p.String()]
		if len(choices) == 0 {
			break
		}
		next := choices[rand.IntN(len(choices))]
		words = append(words, next)
		p.Shift(next)
	}
	return strings.Join(words, " ")
}

func ReadManifesto(numWords int, prefixLen int) string {
	file, err := os.Open("words_manifesto")
	if err != nil {
		_ = fmt.Errorf("%s", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			_ = fmt.Errorf("%s", err)
		}
	}()
	var words []string
	scanner := bufio.NewScanner(file)
	scanner.Scan() // this moves to the next token
	randomSection := rand.IntN(25000)
	endSection := randomSection + numWords
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		if lineCount >= randomSection && lineCount <= endSection {
			words = append(words, scanner.Text())
		}
	}
	stringWord := strings.Join(words, " ")

	return stringWord
}

func BableManifesto(numWords int, prefixLen int) string {
	c := NewChain(prefixLen)
	c.Build(ReadManifesto(numWords, prefixLen))
	text := c.Generate(numWords)

	return text
}
