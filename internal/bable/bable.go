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

// internWord converts a word to its unique ID, creating a new ID if needed.
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
func (chain *Chain) Build(text string) {
	sentences := splitIntoSentences(text)

	for _, sentence := range sentences {
		tokens := tokenize(sentence)
		if len(tokens) == 0 {
			continue
		}

		// Initialize prefix with START tokens
		prefix := make(Prefix, chain.prefixLen)
		startID := chain.internWord(startToken)
		for i := 0; i < chain.prefixLen; i++ {
			prefix[i] = startID
		}

		// Build chain by recording what word follows each prefix
		for _, token := range tokens {
			tokenID := chain.internWord(token)
			key := hashPrefix(prefix)
			chain.chain[key] = append(chain.chain[key], tokenID)
			prefix.Shift(tokenID)
		}

		// Mark sentence end
		endID := chain.internWord(endToken)
		key := hashPrefix(prefix)
		chain.chain[key] = append(chain.chain[key], endID)
	}
}

func splitIntoSentences(text string) []string {
	sentenceDelimiter := regexp.MustCompile(`[.!?]+[\s\n]+`)
	sentences := sentenceDelimiter.Split(text, -1)

	var cleaned []string
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			cleaned = append(cleaned, sentence)
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
	prefix := make(Prefix, chain.prefixLen)

	startID, hasStart := chain.wordToID[startToken]
	if !hasStart {
		return ""
	}

	for i := 0; i < chain.prefixLen; i++ {
		prefix[i] = startID
	}

	var tokens []string
	maxTokens := 100

	for len(tokens) < maxTokens {
		key := hashPrefix(prefix)
		choices := chain.chain[key]

		if len(choices) == 0 {
			break
		}

		nextID := choices[rand.IntN(len(choices))] //nolint:gosec
		nextToken := chain.vocab[nextID]

		if nextToken == endToken || nextToken == startToken {
			break
		}

		tokens = append(tokens, nextToken)
		prefix.Shift(nextID)
	}

	return joinTokens(tokens)
}

// joinTokens reassembles tokens into text, attaching punctuation
// directly to the preceding word without a space.
func joinTokens(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString(tokens[0])

	for _, token := range tokens[1:] {
		if !isPunctuation(token) {
			builder.WriteByte(' ')
		}
		builder.WriteString(token)
	}

	return builder.String()
}

// Shift removes the first element and appends wordID to the end.
func (prefix Prefix) Shift(wordID int) {
	copy(prefix, prefix[1:])
	prefix[len(prefix)-1] = wordID
}

// hashPrefix uses FNV-1a hash algorithm to create a unique key from word IDs.
func hashPrefix(wordIDs []int) uint64 {
	const fnvOffsetBasis uint64 = 14695981039346656037
	const fnvPrime = 1099511628211

	hash := fnvOffsetBasis
	for _, id := range wordIDs {
		hash ^= uint64(id) //nolint:gosec
		hash *= fnvPrime
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
	chain := NewChain(prefixLen)
	chain.Build(ReadManifesto())

	text := chain.GenerateSentences(numSentences)

	return text
}
