// Package pages has all the logic for serving the pages.
package pages

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"Erebus/internal/bable"
	cache "Erebus/internal/rediscache"
)

// GenerateHandler serves dynamically generated tarpit pages.
func GenerateHandler(w http.ResponseWriter, r *http.Request) {

	// Store real-ip in Redis, we will use this as a unique ID
	// later on we check how long this value has been in our memory
	// This will give a better idea of how long some scrapers have been stuck
	if err := cache.SetIP(r); err != nil {
		log.Printf("failed to store IP in cache: %s", err.Error())
		// Continue serving the page even if cache fails
	}

	generatedText := bable.Bable(50, 5)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported by server",
			http.StatusInternalServerError)
		return
	}
	// Set headers to prevent timeouts and caching
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("X-Accel-Buffering", "no") // Disable Nginx buffering
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Connection", "keep-alive")

	// Build a multi-word title from the generated text
	titleWords := strings.Fields(generatedText)
	titleLen := 3 + rand.IntN(4) //nolint:gosec
	if titleLen > len(titleWords) {
		titleLen = len(titleWords)
	}
	title := strings.Join(titleWords[:titleLen], " ")

	// Generate page metadata
	meta := GenerateMeta(title, generatedText, r.URL.Path)

	ts, err := template.ParseFiles("./html/pages/manifest.tmpl")
	if err != nil {
		log.Printf("error reading template: %s", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title          string
		MetaHTML       template.HTML
		NavHTML        template.HTML
		BreadcrumbHTML template.HTML
		BylineHTML     template.HTML
	}{
		Title:          title,
		MetaHTML:       template.HTML(meta.RenderHead()),                                  //nolint:gosec
		NavHTML:        template.HTML(RenderNav(GenerateNavLinks())),                      //nolint:gosec
		BreadcrumbHTML: template.HTML(RenderBreadcrumbs(GenerateBreadcrumbs(r.URL.Path))), //nolint:gosec
		BylineHTML:     template.HTML(RenderByline(meta)),                                 //nolint:gosec
	}

	err = ts.Execute(w, data)
	if err != nil {
		log.Printf("error executing template: %s", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	flusher.Flush()

	// Stream main content slowly
	streamWords(w, flusher, r, strings.Fields(generatedText), 8)

	// Close the streamed paragraph and text div
	_, _ = fmt.Fprint(w, `</p></div>`)
	flusher.Flush()

	// Generate and write sub-sections with headings
	sectionCount := 2 + rand.IntN(3) //nolint:gosec
	sections := GenerateSections(sectionCount)
	_, _ = fmt.Fprint(w, RenderSections(sections))
	flusher.Flush()

	// Article links
	articleLinks := GenerateLinks(8 + rand.IntN(5)) //nolint:gosec
	_, _ = fmt.Fprint(w, `<ul class="article-links">`)
	for _, l := range articleLinks {
		_, _ = fmt.Fprintf(w, `<li><a href="%s">%s</a></li>`,
			l.URL, html.EscapeString(l.Text))
	}
	_, _ = fmt.Fprint(w, `</ul>`)
	flusher.Flush()

	// Pagination
	basePath := r.URL.Path
	if basePath == "/" {
		basePath = "/articles"
	}
	pagination := GeneratePaginationLinks(basePath)
	_, _ = fmt.Fprint(w, `<nav class="pagination">`)
	for _, p := range pagination {
		_, _ = fmt.Fprintf(w, `<a href="%s">%s</a>`, p.URL, html.EscapeString(p.Text))
	}
	_, _ = fmt.Fprint(w, `</nav>`)

	// Close content div
	_, _ = fmt.Fprint(w, `</div>`)
	flusher.Flush()

	// Sidebar
	sidebarLinks := GenerateLinks(5 + rand.IntN(3)) //nolint:gosec
	_, _ = fmt.Fprint(w, RenderSidebar(sidebarLinks))

	// Close layout div
	_, _ = fmt.Fprint(w, `</div>`)
	flusher.Flush()

	// Footer
	footerLinks := GenerateLinks(8 + rand.IntN(4)) //nolint:gosec
	_, _ = fmt.Fprint(w, RenderFooter(footerLinks))

	_, _ = fmt.Fprint(w, `</body></html>`)
	flusher.Flush()
}

func streamWords(w http.ResponseWriter, flusher http.Flusher, r *http.Request, words []string, intervalSeconds float64) {
	i := 0
	for i < len(words) {
		select {
		case <-r.Context().Done():
			return
		default:
			chunkSize := 1 + rand.IntN(8) //nolint:gosec
			if i+chunkSize > len(words) {
				chunkSize = len(words) - i
			}

			chunk := words[i : i+chunkSize]
			_, _ = fmt.Fprint(w, html.EscapeString(strings.Join(chunk, " "))+" ")
			flusher.Flush()

			i += chunkSize

			var delay time.Duration
			if intervalSeconds > 0 {
				delay = time.Duration(intervalSeconds * float64(time.Second))
			} else {
				// Use random delays if no interval is set
				if rand.Float32() < 0.15 { //nolint:gosec
					delay = time.Duration(300+rand.IntN(200)) * time.Millisecond //nolint:gosec
				} else {
					delay = time.Duration(20+rand.IntN(180)) * time.Millisecond //nolint:gosec
				}
			}

			if i < len(words) {
				time.Sleep(delay)
			}
		}
	}
}
