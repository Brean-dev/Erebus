package pages

import (
	"fmt"
	"net/http"
)

// RobotsHandler serves a robots.txt with strategic disallows to bait scrapers.
// Scrapers often chase disallowed paths looking for sensitive content.
func RobotsHandler(w http.ResponseWriter, r *http.Request) {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	host := r.Host
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, `User-agent: *
Disallow: /admin/
Disallow: /admin/dashboard/
Disallow: /private/
Disallow: /internal-documents/
Disallow: /confidential/
Disallow: /api/v1/users/
Disallow: /api/v2/accounts/
Disallow: /backup/
Disallow: /database-exports/
Disallow: /financial-reports/
Disallow: /employee-records/
Disallow: /staging/
Disallow: /debug/
Disallow: /config/
Allow: /

Sitemap: %s/sitemap.xml
`, baseURL)
}
