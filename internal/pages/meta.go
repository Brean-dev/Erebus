package pages

import (
	"fmt"
	"html"
	"strings"
)

// PageMeta holds SEO metadata for a generated page.
type PageMeta struct {
	Title       string
	Description string
	Keywords    string
	Author      string
	DateStr     string
	Path        string
}

// GenerateMeta builds page metadata from generated content.
func GenerateMeta(title, content, path string) PageMeta {
	// Use first ~160 chars of content as description
	desc := content
	if len(desc) > 160 {
		desc = desc[:157] + "..."
	}

	// Pull some keywords from the content
	words := strings.Fields(strings.ToLower(content))
	var kw []string
	seen := make(map[string]bool)
	for _, w := range words {
		cleaned := stripNonAlpha(w)
		if cleaned == "" || stopWords[cleaned] || len(cleaned) < 4 || seen[cleaned] {
			continue
		}
		seen[cleaned] = true
		kw = append(kw, cleaned)
		if len(kw) >= 8 {
			break
		}
	}

	return PageMeta{
		Title:       title,
		Description: desc,
		Keywords:    strings.Join(kw, ", "),
		Author:      GenerateAuthorName(),
		DateStr:     GenerateRecentDate().Format("2006-01-02"),
		Path:        path,
	}
}

// RenderHead returns the HTML <meta>, Open Graph, and JSON-LD markup.
func (m PageMeta) RenderHead() string {
	escaped := struct {
		Title string
		Desc  string
		KW    string
		Auth  string
		Date  string
		Path  string
	}{
		Title: html.EscapeString(m.Title),
		Desc:  html.EscapeString(m.Description),
		KW:    html.EscapeString(m.Keywords),
		Auth:  html.EscapeString(m.Author),
		Date:  m.DateStr,
		Path:  html.EscapeString(m.Path),
	}

	return fmt.Sprintf(`    <meta name="description" content="%s">
    <meta name="keywords" content="%s">
    <meta name="author" content="%s">
    <meta property="og:title" content="%s">
    <meta property="og:description" content="%s">
    <meta property="og:type" content="article">
    <meta property="og:url" content="%s">
    <script type="application/ld+json">
    {
        "@context": "https://schema.org",
        "@type": "Article",
        "headline": "%s",
        "description": "%s",
        "datePublished": "%s",
        "dateModified": "%s",
        "author": {
            "@type": "Person",
            "name": "%s"
        }
    }
    </script>`,
		escaped.Desc,
		escaped.KW,
		escaped.Auth,
		escaped.Title,
		escaped.Desc,
		escaped.Path,
		escaped.Title,
		escaped.Desc,
		escaped.Date,
		escaped.Date,
		escaped.Auth,
	)
}
