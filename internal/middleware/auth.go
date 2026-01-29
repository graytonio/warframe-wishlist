package middleware

import (
	"context"
	"crypto/ecdsa"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/graytonio/warframe-wishlist/pkg/logger"
	"github.com/graytonio/warframe-wishlist/pkg/response"
)

type contextKey string

const UserIDKey contextKey = "userID"

type AuthMiddleware struct {
	jwtPublicKey *ecdsa.PublicKey
}

func NewAuthMiddleware(jwtPublicKey *ecdsa.PublicKey) *AuthMiddleware {
	return &AuthMiddleware{jwtPublicKey: jwtPublicKey}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Debug(ctx, "authenticating request")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn(ctx, "authentication failed: missing authorization header")
			response.Error(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			logger.Warn(ctx, "authentication failed: invalid authorization header format")
			response.Error(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		tokenString := parts[1]
		logger.Debug(ctx, "parsing JWT token")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.jwtPublicKey, nil
		})

		if err != nil || !token.Valid {
			logger.Warn(ctx, "authentication failed: invalid token", "error", err)
			response.Error(w, http.StatusUnauthorized, "invalid token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Warn(ctx, "authentication failed: invalid token claims")
			response.Error(w, http.StatusUnauthorized, "invalid token claims")
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			logger.Warn(ctx, "authentication failed: missing user ID in token")
			response.Error(w, http.StatusUnauthorized, "missing user ID in token")
			return
		}

		logger.Debug(ctx, "authentication successful", "userID", sub)

		// Add userID to both the standard context key and the logger context
		ctx = context.WithValue(ctx, UserIDKey, sub)
		ctx = logger.ContextWithUserID(ctx, sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}
