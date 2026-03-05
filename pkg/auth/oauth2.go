package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// OAuth2Provider defines an OAuth2 authentication provider.
type OAuth2Provider struct {
	config      *oauth2.Config
	providerURL string
	name        string
}

// OAuth2Config defines OAuth2 configuration.
type OAuth2Config struct {
	Provider     string   `yaml:"provider"` // google, github, custom
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url"`
	Scopes       []string `yaml:"scopes"`
	IssuerURL    string   `yaml:"issuer_url,omitempty"` // For OIDC
}

// NewOAuth2Provider creates a new OAuth2 provider.
func NewOAuth2Provider(config *OAuth2Config) (*OAuth2Provider, error) {
	oauth2Cfg := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
	}

	// Set endpoint based on provider
	switch config.Provider {
	case "google":
		oauth2Cfg.Endpoint = oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		}
		if len(oauth2Cfg.Scopes) == 0 {
			oauth2Cfg.Scopes = []string{"openid", "profile", "email"}
		}

	case "github":
		oauth2Cfg.Endpoint = oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		}
		if len(oauth2Cfg.Scopes) == 0 {
			oauth2Cfg.Scopes = []string{"read:user", "user:email"}
		}

	case "oidc":
		if config.IssuerURL == "" {
			return nil, fmt.Errorf("issuer_url required for OIDC provider")
		}
		// OIDC discovery would happen here
		oauth2Cfg.Endpoint = oauth2.Endpoint{
			AuthURL:  config.IssuerURL + "/authorize",
			TokenURL: config.IssuerURL + "/token",
		}

	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}

	return &OAuth2Provider{
		config:      oauth2Cfg,
		providerURL: config.IssuerURL,
		name:        config.Provider,
	}, nil
}

// GetAuthURL generates the OAuth2 authorization URL.
func (p *OAuth2Provider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges authorization code for tokens.
func (p *OAuth2Provider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	return token, nil
}

// GetUserInfo retrieves user information from the provider.
func (p *OAuth2Provider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := p.config.Client(ctx, token)

	var userInfoURL string
	switch p.name {
	case "google":
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "github":
		userInfoURL = "https://api.github.com/user"
	case "oidc":
		userInfoURL = p.providerURL + "/userinfo"
	}

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}
	defer resp.Body.Close()

	// Parse user info (simplified)
	userInfo := &UserInfo{}
	// Would decode JSON response here

	return userInfo, nil
}

// RefreshToken refreshes an expired token.
func (p *OAuth2Provider) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	tokenSource := p.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	return newToken, nil
}

// UserInfo represents authenticated user information.
type UserInfo struct {
	ID            string         `json:"id"`
	Email         string         `json:"email"`
	Name          string         `json:"name"`
	Picture       string         `json:"picture,omitempty"`
	EmailVerified bool           `json:"email_verified"`
	Provider      string         `json:"provider"`
	CreatedAt     time.Time      `json:"created_at"`
	Entitlements  []string       `json:"entitlements,omitempty"` // Roles, groups, permissions from IdP
	Groups        []string       `json:"groups,omitempty"`
	RawClaims     map[string]any `json:"-"` // Original claims for custom mapping
}

// GenerateState generates a secure random state parameter.
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ValidateState validates the OAuth2 state parameter.
func ValidateState(expected, actual string) bool {
	return expected != "" && expected == actual
}

// AuthMiddleware provides HTTP authentication middleware.
func AuthMiddleware(sessionManager *SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract session from cookie or header
			sessionID, err := extractSessionID(r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate session
			session, err := sessionManager.GetSession(r.Context(), sessionID)
			if err != nil || !session.Valid() {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), "user", session.UserInfo)
			ctx = context.WithValue(ctx, "session", session)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractSessionID extracts session ID from request.
func extractSessionID(r *http.Request) (string, error) {
	// Try cookie first
	cookie, err := r.Cookie("session_id")
	if err == nil {
		return cookie.Value, nil
	}

	// Try Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Parse "Bearer <token>"
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			return authHeader[7:], nil
		}
	}

	return "", fmt.Errorf("no session ID found")
}

// GetUserFromContext extracts user info from request context.
func GetUserFromContext(ctx context.Context) (*UserInfo, bool) {
	user, ok := ctx.Value("user").(*UserInfo)
	return user, ok
}
