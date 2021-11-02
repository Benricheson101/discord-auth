package main

import (
	"fmt"
	"net/http"

	"github.com/benricheson101/discord-status/routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/joho/godotenv"
)

var tokenAuth = jwtauth.New("HS256", []byte("owo wats dis"), nil)

func main() {
	godotenv.Load("./.env")
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"uid": 123})
	fmt.Printf("DEBUG: a sample jwt is %s\n", tokenString)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Mount("/auth", routes.OauthRoutes{}.Routes())
	r.Mount("/admin", routes.AdminRoutes{}.Routes())

	http.ListenAndServe(":3333", r)
}
