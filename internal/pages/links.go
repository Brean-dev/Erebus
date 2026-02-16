package pages

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
	"unicode"

	"Erebus/internal/bable"
)

// Link represents a hyperlink on a generated page.
type Link struct {
	URL  string
	Text string
}

var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true,
	"of": true, "in": true, "to": true, "for": true, "is": true,
	"it": true, "by": true, "on": true, "at": true, "as": true,
	"its": true, "was": true, "are": true, "be": true, "has": true,
	"had": true, "have": true, "with": true, "from": true, "this": true,
	"that": true, "which": true, "but": true, "not": true, "all": true,
}

var categories = []string{
	"articles", "blog", "news", "analysis",
	"research", "reports", "archive", "documents",
}

var months = []string{
	"january", "february", "march", "april",
	"may", "june", "july", "august",
	"september", "october", "november", "december",
}

var firstNames = []string{
	"james", "mary", "robert", "patricia", "john",
	"jennifer", "michael", "linda", "david", "elizabeth",
	"william", "barbara", "richard", "susan", "joseph",
	"thomas", "margaret", "charles", "dorothy", "daniel",
}

var lastNames = []string{
	"smith", "johnson", "williams", "brown", "jones",
	"garcia", "miller", "davis", "rodriguez", "martinez",
	"wilson", "anderson", "taylor", "thomas", "moore",
	"jackson", "martin", "lee", "thompson", "white",
}

// GenerateSlug produces a URL-friendly slug from Markov-generated words.
func GenerateSlug(wordCount int) string {
	raw := bable.Bable(1, 1)
	words := strings.Fields(strings.ToLower(raw))

	var kept []string
	for _, w := range words {
		cleaned := stripNonAlpha(w)
		if cleaned == "" || stopWords[cleaned] {
			continue
		}
		kept = append(kept, cleaned)
		if len(kept) >= wordCount {
			break
		}
	}

	if len(kept) == 0 {
		return "page"
	}
	return strings.Join(kept, "-")
}

// titleCase capitalises the first letter of each word.
func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			runes := []rune(w)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

func stripNonAlpha(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// GenerateAuthorName produces a realistic author name.
func GenerateAuthorName() string {
	first := firstNames[rand.IntN(len(firstNames))] //nolint:gosec
	last := lastNames[rand.IntN(len(lastNames))]   //nolint:gosec
	return titleCase(first) + " " + titleCase(last)
}

// GenerateRecentDate returns a date within the past 60 days.
func GenerateRecentDate() time.Time {
	daysAgo := rand.IntN(60) //nolint:gosec
	return time.Now().AddDate(0, 0, -daysAgo)
}

// RandomCategory picks a random category slug.
func RandomCategory() string {
	return categories[rand.IntN(len(categories))] //nolint:gosec
}

// RandomMonth picks a random month name.
func RandomMonth() string {
	return months[rand.IntN(len(months))] //nolint:gosec
}

// GenerateLinks produces a set of links with realistic URL patterns.
func GenerateLinks(count int) []Link {
	links := make([]Link, 0, count)
	for range count {
		links = append(links, generateOneLink())
	}
	return links
}

func generateOneLink() Link {
	slug := GenerateSlug(3 + rand.IntN(2))     //nolint:gosec
	text := bable.Bable(1, 1)
	year := 2023 + rand.IntN(3)               //nolint:gosec

	patterns := []func() Link{
		func() Link {
			cat := RandomCategory()
			return Link{
				URL:  fmt.Sprintf("/%s/%s", cat, slug),
				Text: text,
			}
		},
		func() Link {
			return Link{
				URL:  fmt.Sprintf("/articles/%d/%02d/%s", year, 1+rand.IntN(12), slug), //nolint:gosec
				Text: text,
			}
		},
		func() Link {
			return Link{
				URL:  fmt.Sprintf("/blog/post/%s", slug),
				Text: text,
			}
		},
		func() Link {
			word := GenerateSlug(1)
			return Link{
				URL:  fmt.Sprintf("/tag/%s", word),
				Text: text,
			}
		},
		func() Link {
			name := strings.ToLower(strings.ReplaceAll(GenerateAuthorName(), " ", "-"))
			return Link{
				URL:  fmt.Sprintf("/author/%s", name),
				Text: text,
			}
		},
		func() Link {
			return Link{
				URL:  fmt.Sprintf("/archive/%d/%s", year, RandomMonth()),
				Text: text,
			}
		},
	}

	return patterns[rand.IntN(len(patterns))]() //nolint:gosec
}

// GenerateNavLinks returns fixed-looking navigation category links.
func GenerateNavLinks() []Link {
	nav := make([]Link, len(categories))
	for i, cat := range categories {
		nav[i] = Link{
			URL:  "/" + cat,
			Text: titleCase(cat),
		}
	}
	return nav
}

// GeneratePaginationLinks creates prev/next and page number links.
func GeneratePaginationLinks(basePath string) []Link {
	totalPages := 5 + rand.IntN(20)           //nolint:gosec
	currentPage := 1 + rand.IntN(totalPages)  //nolint:gosec

	var links []Link
	if currentPage > 1 {
		links = append(links, Link{
			URL:  fmt.Sprintf("%s?page=%d", basePath, currentPage-1),
			Text: "Previous",
		})
	}

	start := max(1, currentPage-2)
	end := min(totalPages, currentPage+2)
	for p := start; p <= end; p++ {
		links = append(links, Link{
			URL:  fmt.Sprintf("%s?page=%d", basePath, p),
			Text: fmt.Sprintf("%d", p),
		})
	}

	if currentPage < totalPages {
		links = append(links, Link{
			URL:  fmt.Sprintf("%s?page=%d", basePath, currentPage+1),
			Text: "Next",
		})
	}
	return links
}
