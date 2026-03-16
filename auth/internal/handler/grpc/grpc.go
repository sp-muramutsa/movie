package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"movieexample.com/gen"
)

// SecretProviders defines a provider of secrets for our handler.
type SecretProvider func() []byte

// Handler defines an auth gRPC handler.
type Handler struct {
	secretProvider SecretProvider
	gen.UnimplementedAuthServiceServer
}

// New creates a new auth gRPC handler.
func New(secretProvider SecretProvider) *Handler {
	return &Handler{secretProvider: secretProvider}
}

// GetToken perfomrs verification of user credentials and returns a JWT token in case of success.
func (h *Handler) GetToken(ctx context.Context, req *gen.GetTokenRequest) (*gen.GetTokenResponse, error) {
	username, password := req.Username, req.Password
	if !validCredentials(username, password) {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"iat":      time.Now().Unix(),
		})
	tokenString, err := token.SignedString(h.secretProvider())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &gen.GetTokenResponse{Token: tokenString}, nil
}

// Dummy method to verify credentials.
// TODO: Create a user database table and connect this method for credential verification
func validCredentials(username, password string) bool {
	if username == "" || password == "" {
		return false
	}
	return true
}

// ValidateToken perofmrs JWT token validation.
func (h *Handler) ValidateToken(ctx context.Context, req *gen.ValidateTokenRequest) (*gen.ValidateTokenResponse, error) {
	token, err := jwt.Parse(
		req.Token,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return h.secretProvider(), nil
		})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	var username string
	if v, ok := claims["username"]; ok {
		if u, ok := v.(string); ok {
			username = u
		}
	}

	return &gen.ValidateTokenResponse{Username: username}, nil
}
