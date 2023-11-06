package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

func (app *application) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		if authHeader == "" {
			app.clientError(w, http.StatusUnauthorized)
		}

		auth := strings.SplitN(authHeader, " ", 2)
		if len(auth) != 2 || auth[0] != "Basic" {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		payload, err := base64.StdEncoding.DecodeString(auth[1])
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 || pair[0] != app.devOption.UserName || pair[1] != app.devOption.Password {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)
		app.logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
