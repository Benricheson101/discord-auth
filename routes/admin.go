package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/benricheson101/discord-status/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type AdminRoutes struct{}

func (rs AdminRoutes) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(util.VerifyJWT)

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			token, _, _ := jwtauth.FromContext(r.Context())

			isValid := token.Expiration().After(time.Now())

			fmt.Println("isValid=", isValid)

			t, _ := token.Get("access_token")
			user, _ := getUser(t.(string))

			w.Write([]byte(fmt.Sprintf("protected area. hi %v#%v", user.Username, user.Discriminator)))
		})
	})

	return r
}
