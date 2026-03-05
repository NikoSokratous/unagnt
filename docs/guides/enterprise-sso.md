# Enterprise SSO (SAML / OIDC)

Unagnt supports enterprise identity federation for the Web UI and API via OAuth2, OpenID Connect (OIDC), and SAML 2.0.

## Overview

- **OAuth2**: Google, GitHub, custom providers
- **OIDC**: Any OpenID Connect–compliant IdP (Okta, Auth0, Keycloak, Azure AD)
- **SAML 2.0**: Enterprise IdPs (Okta, Azure AD, OneLogin, etc.)

User entitlements (roles, groups) from the IdP are extracted and stored in `UserInfo` for RBAC and feature gating.

## Auth Configuration Schema

Configure auth in your agent or server config:

```yaml
auth:
  enabled: true
  providers:
    - id: okta-oidc
      type: oidc
      priority: 0
      oidc:
        issuer_url: https://dev-12345.okta.com/oauth2/default
        client_id: your-client-id
        client_secret: ${OKTA_CLIENT_SECRET}  # Use secrets management in prod
        redirect_url: https://app.example.com/auth/callback
        scopes: [openid, profile, email]
        claim_mappings:
          groups: groups

    - id: corporate-saml
      type: saml
      priority: 1
      saml:
        idp_metadata_url: https://idp.corp.com/metadata
        # OR idp_metadata_path: /config/saml/idp-metadata.xml
        entity_id: https://app.example.com/saml
        acs_url: https://app.example.com/saml/acs
        certificate_path: /certs/sp-cert.pem
        key_path: /certs/sp-key.pem
        attribute_mappings:
          groups: groups
          role: entitlements

  session:
    duration: 24h
    cookie_name: agent_session
    cookie_secure: true
    cookie_same_site: Lax

  entitlements:
    claim_names: [groups, roles]
    mapping:
      admin: admin
      developer: developer
```

## OIDC Setup

1. Register an application in your IdP (Okta, Auth0, Keycloak).
2. Set redirect URI to `https://your-app/auth/callback`.
3. Add scopes: `openid`, `profile`, `email`. Add `groups` or `roles` if supported.
4. Copy client ID and secret into config (use secrets management in production).
5. Set `issuer_url` (e.g. `https://dev-12345.okta.com/oauth2/default`).

## SAML Setup

1. **SP metadata**: Generate service provider metadata (or use the entity ID and ACS URL).

2. **IdP configuration**: Add a SAML app in your IdP:
   - Entity ID: `https://your-app.example.com/saml`
   - ACS URL: `https://your-app.example.com/saml/acs`
   - Audience: same as Entity ID

3. **IdP metadata**: Download IdP metadata XML and either:
   - Serve it at a URL and set `idp_metadata_url`, or
   - Place it on disk and set `idp_metadata_path`

4. **Certificates**: Create an SP certificate and key for signing/encryption:
   ```bash
   openssl req -x509 -newkey rsa:2048 -keyout sp-key.pem -out sp-cert.pem -days 365 -nodes
   ```

5. Map SAML attributes to user fields (e.g. `groups`, `roles`, `memberOf`).

## Example IdP Setup (Keycloak)

1. Create a realm and client (confidential).
2. Valid redirect URI: `https://your-app/*`
3. Web origins: `https://your-app`
4. Client authentication: ON
5. Add mapper for groups: Protocol Mapper → User Attribute → `groups` → Token Claim Name `groups`

## Example IdP Setup (Okta)

1. Create an OIDC Web Application.
2. Sign-in redirect: `https://your-app/auth/callback`
3. Assign groups to the app and enable group claims in the token.
4. For SAML: Create a SAML 2.0 app, configure Attribute Statements (e.g. `groups`).

## Entitlements

- **OIDC**: Claims `groups` and `roles` are mapped to `UserInfo.Entitlements` and `UserInfo.Groups`.
- **SAML**: Attributes such as `groups`, `roles`, `memberOf` are mapped to entitlements.
- Use `EntitlementsConfig.claim_names` and `mapping` to align IdP values with internal roles.

## Security Notes

- Never store client secrets in plaintext; use `pkg/secrets` (Vault, AWS Secrets Manager) in production.
- Use `cookie_secure: true` and HTTPS in production.
- Keep SP keys and certificates secure; rotate periodically.
- Use short session durations and refresh tokens where supported.
