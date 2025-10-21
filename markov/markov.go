// Package markov generates random text using Markov chains
package markov

import (
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/tjarratt/babble"
)

// Chain represents a Markov chain for text generation
type Chain struct {
	order int
	chain map[string][]string
}

var (
	defaultChain *Chain
	initialized  bool
)

// InitFromBabbler initializes the Markov chain using a Babbler instance
// It randomly samples from the Babbler's dictionary to create diverse training text
func InitFromBabbler(b babble.Babbler, order int, sampleSize int) error {
	rand.Seed(time.Now().UnixNano())

	// Randomly sample from the Babbler's word list
	words := randomSample(b.Words, sampleSize)

	sep := b.Separator
	if sep == "" {
		sep = " "
	}

	trainingText := strings.Join(words, sep)
	Init(trainingText, order)
	return nil
}

// randomSample randomly selects n items from a slice without replacement
func randomSample(words []string, n int) []string {
	if n > len(words) {
		n = len(words)
	}

	// Fisher-Yates shuffle
	shuffled := make([]string, len(words))
	copy(shuffled, words)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:n]
}

// InitFromReader initializes the Markov chain from an io.Reader
func InitFromReader(r io.Reader, order int) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	Init(string(b), order)
	return nil
}

// Init initializes the default Markov chain with the given order
// Call this once at startup with your training text
func Init(trainingText string, order int) {
	rand.Seed(time.Now().UnixNano())
	defaultChain = NewChain(order)
	defaultChain.Build(trainingText)
	initialized = true
}

// NewChain creates a new Markov chain with a given order
func NewChain(order int) *Chain {
	return &Chain{
		order: order,
		chain: make(map[string][]string),
	}
}

// Build constructs the Markov chain from input text
func (c *Chain) Build(text string) {
	words := strings.Fields(text)

	for i := 0; i < len(words)-c.order; i++ {
		key := strings.Join(words[i:i+c.order], " ")
		next := words[i+c.order]
		c.chain[key] = append(c.chain[key], next)
	}
}

// BuildChain generates text with the specified word count
// This is the main exported function to use as a module
func BuildChain(wordCount int) string {
	if !initialized || defaultChain == nil {
		return ""
	}
	return defaultChain.Generate(wordCount)
}

// Generate creates new text based on the Markov chain
func (c *Chain) Generate(maxWords int) string {
	// Get a random starting key
	keys := make([]string, 0, len(c.chain))
	for k := range c.chain {
		keys = append(keys, k)
	}

	if len(keys) == 0 {
		return ""
	}

	// Start with a random key
	currentKey := keys[rand.Intn(len(keys))]
	result := strings.Split(currentKey, " ")

	// Generate words up to maxWords
	for len(result) < maxWords {
		// Get the next words from the chain
		nextWords, exists := c.chain[currentKey]
		if !exists || len(nextWords) == 0 {
			break
		}

		// Pick a random next word
		nextWord := nextWords[rand.Intn(len(nextWords))]
		result = append(result, nextWord)

		// Slide the window
		words := strings.Split(currentKey, " ")
		words = append(words[1:], nextWord)
		currentKey = strings.Join(words, " ")
	}

	return strings.Join(result, " ")
}

// AddText adds more text to the default chain
func AddText(text string) {
	if initialized && defaultChain != nil {
		defaultChain.Build(text)
	}
}

// Reset clears the default Markov chain
func Reset() {
	if defaultChain != nil {
		defaultChain = NewChain(defaultChain.order)
		initialized = false
	}
}
