package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/benricheson101/discord-status/middleware"
	"github.com/benricheson101/discord-status/models"
)

type OauthRoutes struct{}

func (rs OauthRoutes) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/callback", rs.OAuthCallback)

	return r
}

func (rs OauthRoutes) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	if code == "" {
		writeError(w, "missing `code` in request")
		return
	}

	payload := &models.TokenExchangePayload{}
	payload, err := payload.Get(code)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	user, err := getUser(payload.AccessToken)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	token, err := payload.ToJWT()
	if err != nil {
		writeError(w, err.Error())
		return
	}

	tokenCookie := http.Cookie{Name: middleware.AUTH_COOKIE, Value: token, HttpOnly: true, Path: "/"}
	http.SetCookie(w, &tokenCookie)

	w.Write([]byte(fmt.Sprintf("Logged in as %v#%v!", user.Username, user.Discriminator)))
}

func getUser(token string) (*models.DiscordUser, error) {
	req, _ := http.NewRequest("GET", models.DISCORD_API_BASE_URL+"/users/@me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("unable to get user from discord")
	}

	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	var user models.DiscordUser
	json.Unmarshal([]byte(body), &user)
	return &user, nil
}

// func getUserGuilds(payload TokenExchangePayload) *[]DiscordGuild {
// 	req, _ := http.NewRequest("GET", DISCORD_API_BASE_URL+"/users/@me/guilds", nil)
// 	req.Header.Set("Authorization", "Bearer "+payload.AccessToken)
// 	res, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		// TODO
// 		log.Fatalf("Error getting user: %v\n", err)
// 		return nil
// 	}

// 	defer res.Body.Close()

// 	body, _ := ioutil.ReadAll(res.Body)

// 	var guilds []DiscordGuild
// 	json.Unmarshal([]byte(body), &guilds)
// 	return &guilds
// }

func writeError(w http.ResponseWriter, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	jsonPayload, _ := json.Marshal(map[string]string{
		"error": errorMsg,
	})

	w.Write(jsonPayload)
}
