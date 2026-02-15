package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
)

type Chain struct {
	vocab     []string
	wordToID  map[string]int
	chain     map[uint64][]int
	prefixLen int
}

type Prefix []int

func NewChain(prefixLen int) *Chain {
	return &Chain{
		vocab:     make([]string, 0, 1000),
		wordToID:  make(map[string]int),
		chain:     make(map[uint64][]int),
		prefixLen: prefixLen,
	}
}

func (chain *Chain) internWord(word string) int {
	if id, exists := chain.wordToID[word]; exists {
		return id
	}
	id := len(chain.vocab)
	chain.vocab = append(chain.vocab, word)
	chain.wordToID[word] = id
	return id
}

func (chain *Chain) Build(r string) {
	br := strings.NewReader(r)
	p := make(Prefix, chain.prefixLen)

	for {
		var s string
		if _, err := fmt.Fscan(br, &s); err != nil {
			break
		}

		wordID := chain.internWord(s)
		key := hashPrefix(p)
		chain.chain[key] = append(chain.chain[key], wordID)
		p.Shift(wordID)
	}
}

func (c *Chain) Generate(n int) string {
	p := make(Prefix, c.prefixLen)
	var words []string

	for range n {
		key := hashPrefix(p)
		choices := c.chain[key]

		if len(choices) == 0 {
			break
		}

		nextID := choices[rand.IntN(len(choices))]
		nextWord := c.vocab[nextID]
		words = append(words, nextWord)
		p.Shift(nextID)
	}

	return strings.Join(words, " ")
}

func (p Prefix) Shift(wordID int) {
	copy(p, p[1:])
	p[len(p)-1] = wordID
}

func hashPrefix(wordIDs []int) uint64 {
	var hash uint64 = 14695981039346656037
	for _, id := range wordIDs {
		hash ^= uint64(id)
		hash *= 1099511628211
	}
	return hash
}

func ReadManifesto() string {
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

	for scanner.Scan() {
		word := scanner.Text()
		if word != "" {
			words = append(words, word)
		}
	}

	if err := scanner.Err(); err != nil {
		_ = fmt.Errorf("error reading file: %s", err)
	}
	return strings.Join(words, " ")
}

func BableManifesto(numWords int, prefixLen int) string {
	c := NewChain(prefixLen)
	c.Build(ReadManifesto())
	text := c.Generate(numWords)
	return text
}
