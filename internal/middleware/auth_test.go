package middleware

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// generateTestKeyPair creates an ECDSA key pair for testing
func generateTestKeyPair(t *testing.T) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key pair: %v", err)
	}
	return privateKey, &privateKey.PublicKey
}

func createTestToken(privateKey *ecdsa.PrivateKey, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, _ := token.SignedString(privateKey)
	return tokenString
}

func TestAuthMiddleware_Authenticate_ValidToken(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	userID := "user-123"

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := createTestToken(privateKey, claims)

	middleware := NewAuthMiddleware(publicKey)

	var capturedUserID string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if capturedUserID != userID {
		t.Errorf("expected userID '%s', got '%s'", userID, capturedUserID)
	}
}

func TestAuthMiddleware_Authenticate_MissingHeader(t *testing.T) {
	_, publicKey := generateTestKeyPair(t)
	middleware := NewAuthMiddleware(publicKey)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddleware_Authenticate_InvalidHeaderFormat(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "invalid-token"},
		{"wrong prefix", "Basic token123"},
		{"empty token", "Bearer "},
		{"only bearer", "Bearer"},
	}

	_, publicKey := generateTestKeyPair(t)
	middleware := NewAuthMiddleware(publicKey)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("next handler should not be called")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.header)
			rec := httptest.NewRecorder()

			middleware.Authenticate(nextHandler).ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}
		})
	}
}

func TestAuthMiddleware_Authenticate_InvalidToken(t *testing.T) {
	_, publicKey := generateTestKeyPair(t)
	middleware := NewAuthMiddleware(publicKey)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddleware_Authenticate_WrongKey(t *testing.T) {
	// Sign with one key pair
	signingPrivateKey, _ := generateTestKeyPair(t)
	claims := jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := createTestToken(signingPrivateKey, claims)

	// Validate with a different key pair
	_, validationPublicKey := generateTestKeyPair(t)
	middleware := NewAuthMiddleware(validationPublicKey)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddleware_Authenticate_ExpiredToken(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	claims := jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(-time.Hour).Unix(),
	}
	token := createTestToken(privateKey, claims)

	middleware := NewAuthMiddleware(publicKey)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddleware_Authenticate_MissingSubClaim(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := createTestToken(privateKey, claims)

	middleware := NewAuthMiddleware(publicKey)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddleware_Authenticate_EmptySubClaim(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	claims := jwt.MapClaims{
		"sub": "",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := createTestToken(privateKey, claims)

	middleware := NewAuthMiddleware(publicKey)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddleware_Authenticate_CaseInsensitiveBearer(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	claims := jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := createTestToken(privateKey, claims)

	middleware := NewAuthMiddleware(publicKey)

	bearerVariants := []string{"bearer", "Bearer", "BEARER", "BeArEr"}

	for _, bearer := range bearerVariants {
		t.Run(bearer, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", bearer+" "+token)
			rec := httptest.NewRecorder()

			middleware.Authenticate(nextHandler).ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d for bearer '%s', got %d", http.StatusOK, bearer, rec.Code)
			}
		})
	}
}

func TestGetUserID_WithValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "user-123")
	userID := GetUserID(ctx)

	if userID != "user-123" {
		t.Errorf("expected userID 'user-123', got '%s'", userID)
	}
}

func TestGetUserID_WithoutValue(t *testing.T) {
	ctx := context.Background()
	userID := GetUserID(ctx)

	if userID != "" {
		t.Errorf("expected empty userID, got '%s'", userID)
	}
}

func TestGetUserID_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, 123)
	userID := GetUserID(ctx)

	if userID != "" {
		t.Errorf("expected empty userID for wrong type, got '%s'", userID)
	}
}
