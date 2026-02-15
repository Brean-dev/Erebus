package pages

import (
	"fmt"
	"html"
	"math/rand/v2"
	"strings"

	"Erebus/internal/bable"
)

// Breadcrumb represents a single breadcrumb navigation item.
type Breadcrumb struct {
	Text string
	URL  string
}

// Section represents a content sub-section with heading and body.
type Section struct {
	Heading string
	Content string
	Items   []string // optional list items
}

// GenerateBreadcrumbs builds a breadcrumb trail from the URL path.
func GenerateBreadcrumbs(path string) []Breadcrumb {
	crumbs := []Breadcrumb{{Text: "Home", URL: "/"}}

	path = strings.Trim(path, "/")
	if path == "" {
		return crumbs
	}

	parts := strings.Split(path, "/")
	accumulated := ""
	for _, part := range parts {
		accumulated += "/" + part
		label := strings.ReplaceAll(part, "-", " ")
		label = titleCase(label)
		crumbs = append(crumbs, Breadcrumb{Text: label, URL: accumulated})
	}

	return crumbs
}

// RenderBreadcrumbs returns HTML for breadcrumb navigation.
func RenderBreadcrumbs(crumbs []Breadcrumb) string {
	var b strings.Builder
	b.WriteString(`<nav class="breadcrumb" aria-label="breadcrumb">`)
	for i, c := range crumbs {
		if i > 0 {
			b.WriteString(` <span class="sep">/</span> `)
		}
		if i == len(crumbs)-1 {
			b.WriteString(fmt.Sprintf(`<span class="current">%s</span>`, html.EscapeString(c.Text)))
		} else {
			b.WriteString(fmt.Sprintf(`<a href="%s">%s</a>`, c.URL, html.EscapeString(c.Text)))
		}
	}
	b.WriteString(`</nav>`)
	return b.String()
}

// RenderNav returns HTML for the main navigation menu.
func RenderNav(links []Link) string {
	var b strings.Builder
	b.WriteString(`<nav class="main-nav"><ul>`)
	for _, l := range links {
		b.WriteString(fmt.Sprintf(`<li><a href="%s">%s</a></li>`,
			l.URL, html.EscapeString(l.Text)))
	}
	b.WriteString(`</ul></nav>`)
	return b.String()
}

// GenerateSections produces sub-sections with headings and text.
func GenerateSections(count int) []Section {
	sections := make([]Section, 0, count)
	for range count {
		heading := bable.Bable(1, 1)
		// Trim to first few words for a heading
		words := strings.Fields(heading)
		if len(words) > 6 {
			words = words[:6]
		}
		heading = strings.Join(words, " ")

		content := bable.Bable(3+rand.IntN(5), 3)

		var items []string
		// 30% chance of having a list
		if rand.Float32() < 0.3 {
			listLen := 3 + rand.IntN(4)
			for range listLen {
				items = append(items, bable.Bable(1, 2))
			}
		}

		sections = append(sections, Section{
			Heading: heading,
			Content: content,
			Items:   items,
		})
	}
	return sections
}

// RenderSections returns streamed HTML for sub-sections
// (headings, paragraphs, optional lists).
func RenderSections(sections []Section) string {
	var b strings.Builder
	for _, s := range sections {
		b.WriteString(fmt.Sprintf(`<h2>%s</h2>`, html.EscapeString(s.Heading)))
		b.WriteString(fmt.Sprintf(`<p>%s</p>`, html.EscapeString(s.Content)))
		if len(s.Items) > 0 {
			b.WriteString(`<ul class="content-list">`)
			for _, item := range s.Items {
				b.WriteString(fmt.Sprintf(`<li>%s</li>`, html.EscapeString(item)))
			}
			b.WriteString(`</ul>`)
		}
	}
	return b.String()
}

// RenderSidebar returns HTML for a sidebar with related articles.
func RenderSidebar(links []Link) string {
	var b strings.Builder
	b.WriteString(`<aside class="sidebar"><h3>Related Articles</h3><ul>`)
	for _, l := range links {
		b.WriteString(fmt.Sprintf(`<li><a href="%s">%s</a></li>`,
			l.URL, html.EscapeString(l.Text)))
	}
	b.WriteString(`</ul></aside>`)
	return b.String()
}

// RenderFooter returns HTML for the page footer with topic links.
func RenderFooter(links []Link) string {
	var b strings.Builder
	b.WriteString(`<footer class="site-footer"><div class="footer-links"><h4>Popular Topics</h4><ul>`)
	for _, l := range links {
		b.WriteString(fmt.Sprintf(`<li><a href="%s">%s</a></li>`,
			l.URL, html.EscapeString(l.Text)))
	}
	b.WriteString(`</ul></div>`)
	b.WriteString(fmt.Sprintf(`<p class="copyright">%s. All rights reserved.</p>`,
		html.EscapeString(GenerateAuthorName())))
	b.WriteString(`</footer>`)
	return b.String()
}

// RenderByline returns an author/date byline.
func RenderByline(meta PageMeta) string {
	return fmt.Sprintf(`<div class="byline">By <a href="/author/%s">%s</a> | Published %s</div>`,
		strings.ToLower(strings.ReplaceAll(meta.Author, " ", "-")),
		html.EscapeString(meta.Author),
		meta.DateStr,
	)
}
