package auth

import (
	"fmt"
)

// AuthConfig is the top-level auth configuration schema for Web UI and API.
type AuthConfig struct {
	Enabled      bool               `yaml:"enabled"`
	Providers    []AuthProvider     `yaml:"providers"`
	Session      SessionConfig      `yaml:"session"`
	Entitlements EntitlementsConfig `yaml:"entitlements"`
}

// AuthProvider configures a single identity provider.
type AuthProvider struct {
	ID       string `yaml:"id"`
	Type     string `yaml:"type"`     // oauth2, oidc, saml
	Priority int    `yaml:"priority"` // Lower = preferred

	OAuth2 *OAuth2ProviderConfig `yaml:"oauth2,omitempty"`
	OIDC   *OIDCProviderConfig   `yaml:"oidc,omitempty"`
	SAML   *SAMLProviderConfig   `yaml:"saml,omitempty"`
}

// OAuth2ProviderConfig configures OAuth2 (Google, GitHub, custom).
type OAuth2ProviderConfig struct {
	Provider     string   `yaml:"provider"`
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url"`
	Scopes       []string `yaml:"scopes"`
}

// OIDCProviderConfig configures OpenID Connect.
type OIDCProviderConfig struct {
	IssuerURL    string   `yaml:"issuer_url"`
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url"`
	Scopes       []string `yaml:"scopes"`
	// ClaimMappings maps IdP claims to UserInfo fields (e.g. groups -> Entitlements)
	ClaimMappings map[string]string `yaml:"claim_mappings"`
}

// SAMLProviderConfig configures SAML Service Provider.
type SAMLProviderConfig struct {
	// IdP metadata: use one of IdPMetadataURL or IdPMetadataPath
	IdPMetadataURL  string `yaml:"idp_metadata_url"`
	IdPMetadataPath string `yaml:"idp_metadata_path"`
	EntityID        string `yaml:"entity_id"`
	ACSURL          string `yaml:"acs_url"`
	CertificatePath string `yaml:"certificate_path"`
	KeyPath         string `yaml:"key_path"`
	// AttributeMappings maps SAML attribute names to UserInfo fields
	// e.g. groups -> entitlements, role -> entitlements
	AttributeMappings map[string]string `yaml:"attribute_mappings"`
}

// SessionConfig configures session behavior.
type SessionConfig struct {
	Duration       string `yaml:"duration"` // e.g. 24h
	CookieName     string `yaml:"cookie_name"`
	CookieSecure   bool   `yaml:"cookie_secure"`
	CookieSameSite string `yaml:"cookie_same_site"` // Strict, Lax, None
}

// EntitlementsConfig configures how entitlements (roles, groups) are used.
type EntitlementsConfig struct {
	// Roles is a list of role identifiers; users must have at least one from IdP
	Roles []string `yaml:"roles,omitempty"`
	// ClaimNames lists claim/attribute names that provide entitlements (e.g. groups, roles)
	ClaimNames []string `yaml:"claim_names,omitempty"`
	// Mapping maps IdP role/group values to internal role names
	Mapping map[string]string `yaml:"mapping,omitempty"`
}

// Validate checks AuthConfig for required fields.
func (c *AuthConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if len(c.Providers) == 0 {
		return fmt.Errorf("auth: at least one provider required when enabled")
	}
	for i, p := range c.Providers {
		if p.Type == "" {
			return fmt.Errorf("auth: provider %d: type is required", i)
		}
		switch p.Type {
		case "oauth2":
			if p.OAuth2 == nil {
				return fmt.Errorf("auth: provider %s: oauth2 config required", p.ID)
			}
			if p.OAuth2.ClientID == "" || p.OAuth2.ClientSecret == "" || p.OAuth2.RedirectURL == "" {
				return fmt.Errorf("auth: provider %s: oauth2 client_id, client_secret, redirect_url required", p.ID)
			}
		case "oidc":
			if p.OIDC == nil {
				return fmt.Errorf("auth: provider %s: oidc config required", p.ID)
			}
			if p.OIDC.IssuerURL == "" || p.OIDC.ClientID == "" || p.OIDC.ClientSecret == "" || p.OIDC.RedirectURL == "" {
				return fmt.Errorf("auth: provider %s: oidc issuer_url, client_id, client_secret, redirect_url required", p.ID)
			}
		case "saml":
			if p.SAML == nil {
				return fmt.Errorf("auth: provider %s: saml config required", p.ID)
			}
			if (p.SAML.IdPMetadataURL == "" && p.SAML.IdPMetadataPath == "") || p.SAML.ACSURL == "" ||
				p.SAML.CertificatePath == "" || p.SAML.KeyPath == "" {
				return fmt.Errorf("auth: provider %s: saml requires idp_metadata_url or idp_metadata_path, acs_url, certificate_path, key_path", p.ID)
			}
		default:
			return fmt.Errorf("auth: provider %s: unsupported type %q", p.ID, p.Type)
		}
	}
	return nil
}
