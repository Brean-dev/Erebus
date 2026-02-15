package pages

import (
	"encoding/xml"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"
)

type urlSet struct {
	XMLName xml.Name  `xml:"urlset"`
	XMLNS   string    `xml:"xmlns,attr"`
	URLs    []siteURL `xml:"url"`
}

type siteURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq"`
	Priority   string `xml:"priority"`
}

// SitemapHandler dynamically generates a sitemap.xml with realistic URLs.
// Each request produces different URLs, creating an infinite crawl surface.
func SitemapHandler(w http.ResponseWriter, r *http.Request) {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

	urlCount := 50 + rand.IntN(51) // 50-100 URLs
	urls := make([]siteURL, 0, urlCount)

	// Add the homepage
	urls = append(urls, siteURL{
		Loc:        baseURL + "/",
		LastMod:    time.Now().Format("2006-01-02"),
		ChangeFreq: "daily",
		Priority:   "1.0",
	})

	// Add category index pages
	for _, cat := range categories {
		urls = append(urls, siteURL{
			Loc:        fmt.Sprintf("%s/%s", baseURL, cat),
			LastMod:    GenerateRecentDate().Format("2006-01-02"),
			ChangeFreq: "weekly",
			Priority:   "0.9",
		})
	}

	// Fill the rest with generated article URLs
	freqs := []string{"daily", "weekly", "monthly"}
	for len(urls) < urlCount {
		link := generateOneLink()
		priority := fmt.Sprintf("%.1f", 0.5+rand.Float64()*0.4)

		urls = append(urls, siteURL{
			Loc:        baseURL + link.URL,
			LastMod:    GenerateRecentDate().Format("2006-01-02"),
			ChangeFreq: freqs[rand.IntN(len(freqs))],
			Priority:   priority,
		})
	}

	sitemap := urlSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	enc.Encode(sitemap)
}
