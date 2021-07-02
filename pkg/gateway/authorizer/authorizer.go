package authorizer

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/fernandodr19/library/pkg/domain"
	"github.com/fernandodr19/library/pkg/domain/entities/accounts"
	"github.com/fernandodr19/library/pkg/domain/vos"
	"github.com/google/uuid"
)

type BearerAuthorizer struct {
	secretKey []byte
}

func New(secretKey string) (*BearerAuthorizer, error) {
	return &BearerAuthorizer{[]byte(secretKey)}, nil
}

type Payload struct {
	TokenID   uuid.UUID  `json:"token_id"`
	UserID    vos.UserID `json:"user_id"`
	IssuedAt  time.Time  `json:"issued_at"`
	ExpiredAt time.Time  `json:"expired_at"`
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}

func (b *BearerAuthorizer) CreateToken(userID vos.UserID, duration time.Duration) (vos.AccessToken, error) {
	const operation = "authorizer.BearerAuthorizer.CreateToken"

	tokenID, err := uuid.NewRandom()
	if err != nil {
		return "", domain.Error(operation, err)
	}

	now := time.Now()

	payload := &Payload{
		TokenID:   tokenID,
		UserID:    userID,
		IssuedAt:  now,
		ExpiredAt: now.Add(duration),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	token, err := jwtToken.SignedString(b.secretKey)
	if err != nil {
		return "", domain.Error(operation, err)
	}

	return vos.AccessToken(token), nil
}

func (a *BearerAuthorizer) AuthorizeRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// TODO use responses package maybe?
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Missing auth header"))
			return
		}

		splitedAuthHeader := strings.Split(authHeader, " ")
		if len(splitedAuthHeader) != 2 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid auth header"))
			return
		}

		token := splitedAuthHeader[1]

		payload, err := a.verifyToken(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), accounts.UserIDContextKey, payload.UserID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (b *BearerAuthorizer) verifyToken(token string) (*Payload, error) {
	const operation = "authorizer.BearerAuthorizer.VerifyToken"

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, domain.Error(operation, ErrInvalidToken)
		}
		return b.secretKey, nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		return nil, domain.Error(operation, err)
	}

	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, domain.Error(operation, ErrExpiredToken)
		}
		return nil, domain.Error(operation, ErrInvalidToken)
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, domain.Error(operation, ErrInvalidToken)
	}

	return payload, nil
}
