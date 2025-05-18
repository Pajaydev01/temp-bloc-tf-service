package middleware

import (
	"os"

	"github.com/rs/cors"
	"github.com/unrolled/secure"
)

// Cors middleware enables cors
func Cors() *cors.Cors {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
	})
	return c
}

// secure the application
func Secure() *secure.Secure {
	options := secure.Options{
		BrowserXssFilter:     true,
		ContentTypeNosniff:   true,
		FrameDeny:            false,
		SSLForceHost:         false,
		STSIncludeSubdomains: true,
		STSPreload:           true,
	}

	if os.Getenv("ENVIRONMENT") != "production" {
		options.IsDevelopment = true
	}

	secureMiddleware := secure.New(options)

	return secureMiddleware
}
