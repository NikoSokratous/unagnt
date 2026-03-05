package auth

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

// SAMLConfig holds configuration for the SAML Service Provider.
type SAMLConfig struct {
	// IdPMetadataURL is the URL to fetch IdP metadata from (e.g. https://idp.example.com/metadata).
	// Either IdPMetadataURL or IdPMetadataPath must be set.
	IdPMetadataURL string
	// IdPMetadataPath is the local file path to IdP metadata XML.
	// Either IdPMetadataURL or IdPMetadataPath must be set.
	IdPMetadataPath string
	// EntityID is the SAML entity ID of this Service Provider.
	EntityID string
	// ACSURL is the Assertion Consumer Service URL (e.g. https://sp.example.com/saml/acs).
	ACSURL string
	// CertificatePath is the path to the SP X.509 certificate file.
	CertificatePath string
	// KeyPath is the path to the SP private key file.
	KeyPath string
}

// SAMLProvider implements a SAML Service Provider using samlsp.
type SAMLProvider struct {
	config     *SAMLConfig
	middleware *samlsp.Middleware
}

// NewSAMLProvider creates a new SAML Service Provider from the given config.
// IdP metadata is loaded from either IdPMetadataURL (fetched via HTTP) or
// IdPMetadataPath (read from local file). Exactly one must be set.
func NewSAMLProvider(ctx context.Context, config *SAMLConfig) (*SAMLProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("saml: config is required")
	}

	var idpMetadata *saml.EntityDescriptor
	if config.IdPMetadataURL != "" {
		metadataURL, err := url.Parse(config.IdPMetadataURL)
		if err != nil {
			return nil, fmt.Errorf("saml: invalid IdPMetadataURL: %w", err)
		}
		metadata, err := samlsp.FetchMetadata(ctx, http.DefaultClient, *metadataURL)
		if err != nil {
			return nil, fmt.Errorf("saml: fetch IdP metadata from URL: %w", err)
		}
		idpMetadata = metadata
	} else if config.IdPMetadataPath != "" {
		data, err := os.ReadFile(config.IdPMetadataPath)
		if err != nil {
			return nil, fmt.Errorf("saml: read IdP metadata file: %w", err)
		}
		metadata, err := samlsp.ParseMetadata(data)
		if err != nil {
			return nil, fmt.Errorf("saml: parse IdP metadata: %w", err)
		}
		idpMetadata = metadata
	} else {
		return nil, fmt.Errorf("saml: either IdPMetadataURL or IdPMetadataPath must be set")
	}

	acsURL, err := url.Parse(config.ACSURL)
	if err != nil {
		return nil, fmt.Errorf("saml: invalid ACSURL: %w", err)
	}
	rootURL := &url.URL{
		Scheme: acsURL.Scheme,
		Host:   acsURL.Host,
		Path:   "/",
	}

	keyPair, err := tls.LoadX509KeyPair(config.CertificatePath, config.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("saml: load certificate/key: %w", err)
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("saml: parse certificate: %w", err)
	}

	entityID := config.EntityID
	if entityID == "" {
		entityID = config.ACSURL
	}

	opts := samlsp.Options{
		EntityID:    entityID,
		URL:         *rootURL,
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
	}

	middleware, err := samlsp.New(opts)
	if err != nil {
		return nil, fmt.Errorf("saml: create middleware: %w", err)
	}

	return &SAMLProvider{
		config:     config,
		middleware: middleware,
	}, nil
}

// Middleware returns the samlsp.Middleware for HTTP handler integration.
func (p *SAMLProvider) Middleware() *samlsp.Middleware {
	return p.middleware
}

// GetUserInfo extracts UserInfo from a SAML assertion.
func (p *SAMLProvider) GetUserInfo(assertion *saml.Assertion) (*UserInfo, error) {
	if assertion == nil {
		return nil, fmt.Errorf("saml: assertion is nil")
	}

	info := &UserInfo{
		Provider:  "saml",
		CreatedAt: time.Now(),
	}

	if assertion.Subject != nil && assertion.Subject.NameID != nil {
		info.ID = assertion.Subject.NameID.Value
	}

	getAttr := func(names ...string) string {
		for _, attrStmt := range assertion.AttributeStatements {
			for _, attr := range attrStmt.Attributes {
				for _, name := range names {
					if attr.Name == name || attr.FriendlyName == name {
						if len(attr.Values) > 0 {
							return attr.Values[0].Value
						}
						break
					}
				}
			}
		}
		return ""
	}

	info.Email = getAttr(
		"email", "mail", "Email", "Mail",
		"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
	)
	info.Name = getAttr(
		"name", "cn", "displayName", "Name", "CN", "DisplayName",
		"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
	)
	info.Picture = getAttr("picture", "avatar", "thumbnailPhoto")

	// Entitlements from multi-valued attributes (groups, roles, memberOf)
	entitlementNames := []string{
		"groups", "roles", "memberOf", "Groups", "Roles",
		"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role",
	}
	for _, attrStmt := range assertion.AttributeStatements {
		for _, attr := range attrStmt.Attributes {
			for _, en := range entitlementNames {
				if attr.Name == en || attr.FriendlyName == en {
					for _, v := range attr.Values {
						if v.Value != "" {
							info.Entitlements = append(info.Entitlements, v.Value)
							info.Groups = append(info.Groups, v.Value)
						}
					}
					break
				}
			}
		}
	}

	return info, nil
}
