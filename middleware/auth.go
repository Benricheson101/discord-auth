package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/benricheson101/discord-status/models"
	"github.com/golang-jwt/jwt"
)

const (
	AUTH_COOKIE = "token"

	CTX_KEY_AUTH = "auth"
)

func VerifyJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	return token, err
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		tok, err := r.Cookie(AUTH_COOKIE)
		if err != nil {
			tokenString = r.Header.Get("Authorization")

			if tokenString == "" {
				authFailed(w, "missing token in request")
				return
			}
		} else {
			tokenString = tok.Value
		}

		token, err := VerifyJWT(tokenString)
		if err != nil {
			authFailed(w, "failed to verify auth token")
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		exp := int64(claims["exp"].(float64))
		accessToken := claims["t"].(string)
		refreshToken := claims["r"].(string)

		var payload models.TokenExchangePayload
		payload = models.TokenExchangePayload{AccessToken: accessToken, RefreshToken: refreshToken, ExpiresIn: exp - time.Now().UnixMilli()}

		if time.Now().UnixMilli() > int64(exp) {
			refreshedPayload, err := payload.Refresh()
			if err != nil {
				authFailed(w, "failed to refresh token")
				return
			}

			newJwt, err := refreshedPayload.ToJWT()
			if err != nil {
				authFailed(w, "failed to create jwt from token exchange payload")
				return
			}

			tokenCookie := http.Cookie{Name: AUTH_COOKIE, Value: newJwt, HttpOnly: true, Path: "/"}
			http.SetCookie(w, &tokenCookie)
			payload = *refreshedPayload
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, CTX_KEY_AUTH, payload)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func authFailed(w http.ResponseWriter, errorMsg string) {
	http.SetCookie(w, &http.Cookie{Name: AUTH_COOKIE, Path: "/", Value: "", Expires: time.Unix(0, 0)})

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusUnauthorized)

	jsonPayload, _ := json.Marshal(map[string]string{
		"error": errorMsg,
	})

	w.Write(jsonPayload)
}
