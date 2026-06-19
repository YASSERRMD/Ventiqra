package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/YASSERRMD/Ventiqra/backend/internal/auth"
	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

// Public user shape returned by auth endpoints (never includes password_hash).
type publicUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type authResponse struct {
	User      publicUser `json:"user"`
	Token     string     `json:"token"`
	ExpiresAt string     `json:"expires_at"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if s.users == nil || s.tokens == nil {
		writeError(w, http.StatusServiceUnavailable, "auth not configured")
		return
	}

	var req registerRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	if !isEmail(req.Email) {
		writeError(w, http.StatusBadRequest, "a valid email is required")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := s.users.CreateUser(r.Context(), req.Email, hash, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		s.log.Error("create user failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	s.issueAndRespond(w, r, user)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if s.users == nil || s.tokens == nil {
		writeError(w, http.StatusServiceUnavailable, "auth not configured")
		return
	}

	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := s.users.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	s.issueAndRespond(w, r, user)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	if s.users == nil {
		writeError(w, http.StatusServiceUnavailable, "auth not configured")
		return
	}
	userID := middleware.UserIDFrom(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := s.users.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user": publicUser{ID: user.ID, Email: user.Email, Name: user.Name},
	})
}

func (s *Server) issueAndRespond(w http.ResponseWriter, r *http.Request, user *repository.User) {
	token, exp, err := s.tokens.Issue(user.ID)
	if err != nil {
		s.log.Error("issue token failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	writeJSON(w, http.StatusOK, authResponse{
		User:      publicUser{ID: user.ID, Email: user.Email, Name: user.Name},
		Token:     token,
		ExpiresAt: exp.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// tokenParser adapts the TokenManager to the middleware.TokenParser signature.
func (s *Server) tokenParser() middleware.TokenParser {
	return func(token string) (string, error) {
		claims, err := s.tokens.Parse(token)
		if err != nil {
			return "", err
		}
		return claims.UserID, nil
	}
}

// isEmail performs a minimal email sanity check.
func isEmail(s string) bool {
	at := strings.Index(s, "@")
	return at > 0 && strings.Contains(s[at+1:], ".") && !strings.HasSuffix(s, ".")
}
