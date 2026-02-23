// Package bable provides Markov chain text generation functionality.
package bable

import (
	"bufio"
	"encoding/binary"
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
	chain     map[string][]int
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
		chain:     make(map[string][]int),
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
	startID := chain.internWord(startToken)
	endID := chain.internWord(endToken)

	for _, sentence := range sentences {
		words := tokenize(sentence)
		if len(words) == 0 {
			continue
		}

		prefix := make(Prefix, chain.prefixLen)
		for i := range prefix {
			prefix[i] = startID
		}

		for _, word := range words {
			wordID := chain.internWord(word)
			key := encodePrefix(prefix)
			chain.chain[key] = append(chain.chain[key], wordID)
			prefix.Shift(wordID)
		}

		key := encodePrefix(prefix)
		chain.chain[key] = append(chain.chain[key], endID)
	}
}

// sentenceRe matches sentence-ending punctuation followed by whitespace and a capital letter.
// Requiring a capital letter avoids false splits on abbreviations like "i.e." or "Mr.".
var sentenceRe = regexp.MustCompile(`([.!?]+)\s+([A-Z])`)

func splitIntoSentences(text string) []string {
	// Insert a null byte between the punctuation and the capital letter so we can
	// split there without losing the capital that starts the next sentence.
	separated := sentenceRe.ReplaceAllString(text, "$1\x00$2")
	parts := strings.Split(separated, "\x00")

	sentences := make([]string, 0, len(parts))
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if s != "" {
			sentences = append(sentences, s)
		}
	}
	return sentences
}

// wordRe matches individual words only (no punctuation in the chain).
var wordRe = regexp.MustCompile(`[a-zA-Z']+`)

// tokenize splits text into individual words, ignoring punctuation.
func tokenize(text string) []string {
	return wordRe.FindAllString(text, -1)
}

// encodePrefix encodes word IDs as a binary string key, avoiding hash collisions.
func encodePrefix(prefix []int) string {
	buf := make([]byte, len(prefix)*8)
	for i, id := range prefix {
		binary.LittleEndian.PutUint64(buf[i*8:], uint64(id)) //nolint:gosec
	}
	return string(buf)
}

// GenerateSentences produces complete sentences using START/END tokens.
func (chain *Chain) GenerateSentences(numSentences int) string {
	sentences := make([]string, 0, numSentences)

	for range numSentences {
		sentence := chain.generateOneSentence()
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
	}

	return strings.Join(sentences, " ")
}

func (chain *Chain) generateOneSentence() string {
	startID, hasStart := chain.wordToID[startToken]
	if !hasStart {
		return ""
	}

	prefix := make(Prefix, chain.prefixLen)
	for i := range prefix {
		prefix[i] = startID
	}

	tokens := make([]string, 0, 20)
	const maxTokens = 100

	for len(tokens) < maxTokens {
		key := encodePrefix(prefix)
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

	return strings.Join(tokens, " ")
}

// Shift removes the first element and appends wordID to the end.
func (prefix Prefix) Shift(wordID int) {
	copy(prefix, prefix[1:])
	prefix[len(prefix)-1] = wordID
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

	var sb strings.Builder
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			if sb.Len() > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("error reading file: %s\n", err)
	}

	return sb.String()
}

// Bable generates random text by building a Markov chain from the manifesto.
func Bable(numSentences int, prefixLen int) string {
	chain := NewChain(prefixLen)
	chain.Build(ReadManifesto())
	return chain.GenerateSentences(numSentences)
}
