// Package bable provides Markov chain text generation functionality.
package bable

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"regexp"
	"strings"
)

// Chain is a Markov chain text generator.
type Chain struct {
	vocab     []string
	wordToID  map[string]int
	chain     map[uint64][]int
	prefixLen int
}

// Prefix represents a word ID prefix used in the Markov chain.
type Prefix []int

const (
	startToken = "<START>"
	endToken   = "<END>"
)

// NewChain creates a new Markov chain with the specified prefix length.
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

// Build processes text into the Markov chain.
func (chain *Chain) Build(r string) {
	sentences := splitIntoSentences(r)

	for _, sentence := range sentences {
		tokens := tokenize(sentence)
		if len(tokens) == 0 {
			continue
		}

		p := make(Prefix, chain.prefixLen)

		startID := chain.internWord(startToken)
		for i := 0; i < chain.prefixLen; i++ {
			p[i] = startID
		}

		for _, token := range tokens {
			tokenID := chain.internWord(token)
			key := hashPrefix(p)
			chain.chain[key] = append(chain.chain[key], tokenID)
			p.Shift(tokenID)
		}

		endID := chain.internWord(endToken)
		key := hashPrefix(p)
		chain.chain[key] = append(chain.chain[key], endID)
	}
}

func splitIntoSentences(text string) []string {
	// Split on sentence-ending punctuation
	re := regexp.MustCompile(`[.!?]+[\s\n]+`)
	sentences := re.Split(text, -1)

	var cleaned []string
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s != "" {
			cleaned = append(cleaned, s)
		}
	}
	return cleaned
}

var tokenizeRe = regexp.MustCompile(`[\w']+|[,;:\-\(\)\"]+`)

// tokenize splits text into individual words and punctuation tokens.
// Punctuation is separated from words so the chain learns word-level
// transitions and where punctuation naturally appears.
// It's just a regex though.
func tokenize(text string) []string {
	return tokenizeRe.FindAllString(text, -1)
}

// isPunctuation checks if the current string is a punctuation token.
func isPunctuation(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '\'' {
			return false
		}
	}
	return true
}

// GenerateSentences produces complete sentences using START/END tokens.
func (chain *Chain) GenerateSentences(numSentences int) string {
	var sentences []string

	for range numSentences {
		sentence := chain.generateOneSentence()
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
	}

	return strings.Join(sentences, " ")
}

func (chain *Chain) generateOneSentence() string {
	p := make(Prefix, chain.prefixLen)

	startID, hasStart := chain.wordToID[startToken]
	if !hasStart {
		return ""
	}

	for i := 0; i < chain.prefixLen; i++ {
		p[i] = startID
	}

	var tokens []string
	maxTokens := 100

	for len(tokens) < maxTokens {
		key := hashPrefix(p)
		choices := chain.chain[key]

		if len(choices) == 0 {
			break
		}

		nextID := choices[rand.IntN(len(choices))]  //nolint:gosec
		nextToken := chain.vocab[nextID]

		if nextToken == endToken || nextToken == startToken {
			break
		}

		tokens = append(tokens, nextToken)
		p.Shift(nextID)
	}

	return joinTokens(tokens)
}

// joinTokens reassembles tokens into text, attaching punctuation
// directly to the preceding word without a space.
func joinTokens(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(tokens[0])

	for _, token := range tokens[1:] {
		if !isPunctuation(token) {
			b.WriteByte(' ')
		}
		b.WriteString(token)
	}

	return b.String()
}

// Shift removes the first element and appends wordID to the end.
func (p Prefix) Shift(wordID int) {
	copy(p, p[1:])
	p[len(p)-1] = wordID
}

func hashPrefix(wordIDs []int) uint64 {
	var hash uint64 = 14695981039346656037
	for _, id := range wordIDs {
		hash ^= uint64(id) //nolint:gosec
		hash *= 1099511628211
	}
	return hash
}

// ReadManifesto reads the manifest file and returns its contents as a single string.
func ReadManifesto() string {
	file, err := os.Open("manifest")
	if err != nil {
		fmt.Printf("error opening file: %s\n", err)
		return ""
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("error closing file: %s\n", err)
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
		fmt.Printf("error reading file: %s\n", err)
	}

	return strings.Join(words, " ")
}

// Bable generates random text by building a Markov chain from the manifesto.
func Bable(numSentences int, prefixLen int) string {
	c := NewChain(prefixLen)
	c.Build(ReadManifesto())

	text := c.GenerateSentences(numSentences)

	return text
}
