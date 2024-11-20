package middlewares

import (
	"github.com/rs/cors"
	"github.com/unrolled/secure"
	"net/http"
)

func CORSMiddleware(baseURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		c := cors.New(cors.Options{
			AllowedOrigins:   []string{baseURL}, // Use the dynamic base URL
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept"},
			ExposedHeaders:   []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           86400, // Cache preflight response for 24 hours
		})
		return c.Handler(next)
	}
}

// SecurityMiddleware adds HTTP security headers for production
func SecurityMiddleware(next http.Handler) http.Handler {
	secureMiddleware := secure.New(secure.Options{
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
		ForceSTSHeader:       true,
		STSSeconds:           31536000,
		STSIncludeSubdomains: true,
		STSPreload:           true,
		ReferrerPolicy:       "strict-origin-when-cross-origin",
		SSLRedirect:          true,
		IsDevelopment:        true, // Set to false in production
	})
	return secureMiddleware.Handler(next)
}
