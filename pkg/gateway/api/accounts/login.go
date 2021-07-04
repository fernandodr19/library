package accounts

import (
	"encoding/json"
	"net/http"

	"github.com/fernandodr19/library/pkg/domain"
	"github.com/fernandodr19/library/pkg/domain/vos"
	"github.com/fernandodr19/library/pkg/gateway/api/responses"
)

// Login
// @Summary Authenticate user credentials
// @Description Authenticate user credentials.
// @Tags Accounts
// @Param Body body LoginRequest true "Body"
// @Accept json
// @Produce json
// @Success 200 {object} LoginResponse
// @Failure 401 "Invalid password"
// @Failure 404 "User not found"
// @Failure 500 "Internal server error"
// @Router /login [post]
// Login handles login requests
func (h Handler) Login(r *http.Request) responses.Response {
	operation := "accounts.Handler.Login"

	ctx := r.Context()
	var body LoginRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return responses.BadRequest(domain.Error(operation, err), responses.ErrInvalidBody)
	}

	tokens, err := h.Usecase.Login(ctx, body.Email, body.Password)
	if err != nil {
		return responses.ErrorResponse(domain.Error(operation, err))
	}

	return responses.OK(LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// LoginRequest payload
type LoginRequest struct {
	Email    vos.Email    `json:"email"`
	Password vos.Password `json:"password"`
}

// LoginResponse payload
type LoginResponse struct {
	AccessToken  vos.AccessToken  `json:"access_token"`
	RefreshToken vos.RefreshToken `json:"refresh_token"`
}
