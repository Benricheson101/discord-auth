package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	DISCORD_API_BASE_URL = "https://discord.com/api/v9"
	REDIRECT_URI         = "http://localhost:3333/auth/callback"
)

type TokenExchangePayload struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

type DiscordUser struct {
	Id            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	PublicFlags   int    `json:"public_flags"`
	Flags         int    `json:"flags"`
	Banner        string `json:"banner"`
	BannerColor   string `json:"banner_color"`
	Locale        string `json:"locale"`
	MfaEnabled    bool   `json:"mfa_enabled"`
	PremiumType   int    `json:"premium_type"`
	Email         string `json:"email,omitempty"`
	Verified      bool   `json:"verified,omitempty"`
}

type DiscordGuild struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	Permissions string   `json:"permissions"`
	Features    []string `json:"features"`
}

func (t TokenExchangePayload) ToJWT() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"t":   t.AccessToken,
		"r":   t.RefreshToken,
		"exp": time.Now().UnixMilli() + int64(t.ExpiresIn),
	})

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// TODO: somehow make the next 2 functions not 99% identical
func (t *TokenExchangePayload) Get(code string) (*TokenExchangePayload, error) {
	var (
		CLIENT_ID     = os.Getenv("CLIENT_ID")
		CLIENT_SECRET = os.Getenv("CLIENT_SECRET")
	)
	data := url.Values{}

	data.Set("grant_type", "authorization_code")
	data.Set("client_id", CLIENT_ID)
	data.Set("client_secret", CLIENT_SECRET)
	data.Set("code", code)
	data.Set("redirect_uri", REDIRECT_URI)

	req, err := http.NewRequest("POST", DISCORD_API_BASE_URL+"/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to create http request: %v", err))
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New("discord responded with non-200 response code")
	}

	var payload TokenExchangePayload
	json.Unmarshal([]byte(body), &payload)

	t = &payload

	return &payload, nil
}

func (t *TokenExchangePayload) Refresh() (*TokenExchangePayload, error) {
	var (
		CLIENT_ID     = os.Getenv("CLIENT_ID")
		CLIENT_SECRET = os.Getenv("CLIENT_SECRET")
	)
	data := url.Values{}

	data.Set("grant_type", "refresh_token")
	data.Set("client_id", CLIENT_ID)
	data.Set("client_secret", CLIENT_SECRET)
	data.Set("refresh_token", t.RefreshToken)

	req, err := http.NewRequest("POST", DISCORD_API_BASE_URL+"/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to create http request: %v", err))
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New("discord responded with non-200 response code")
	}

	var payload TokenExchangePayload
	json.Unmarshal([]byte(body), &payload)

	t = &payload

	return &payload, nil
}
