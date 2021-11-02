package routes

import (
	"fmt"
	"net/http"

	"github.com/benricheson101/discord-status/middleware"
	"github.com/benricheson101/discord-status/models"

	"github.com/go-chi/chi/v5"
)

type AdminRoutes struct{}

func (rs AdminRoutes) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			token, ok := r.Context().Value(middleware.CTX_KEY_AUTH).(models.TokenExchangePayload)
			if !ok {
				return
			}

			user, _ := getUser(token.AccessToken)

			w.Write([]byte(fmt.Sprintf("protected area. hi %v#%v", user.Username, user.Discriminator)))
		})
	})

	return r
}
