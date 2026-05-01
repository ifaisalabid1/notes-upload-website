package middleware

import "net/http"

func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()

		// Prevent browsers from MIME-sniffing responses away from declared type.
		h.Set("X-Content-Type-Options", "nosniff")

		// Deny embedding this API in any iframe — it's an API, not a page.
		h.Set("X-Frame-Options", "DENY")

		// Basic XSS protection for older browsers.
		h.Set("X-XSS-Protection", "1; mode=block")

		// Only send the origin as referrer, never the full URL.
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Force HTTPS for 1 year once the browser has seen it.
		// Only set in production — breaks local dev with HTTP.
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Restrict what this API's responses can do in a browser context.
		h.Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")

		next.ServeHTTP(w, r)
	})
}
