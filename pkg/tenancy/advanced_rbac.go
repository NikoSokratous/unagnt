package tenancy

import (
	"context"
	"errors"
)

var ErrPermissionDenied = errors.New("permission denied")

// RoleTemplate defines a custom role derived from a base role with extra permissions.
type RoleTemplate struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	BaseRole     Role         `json:"base_role"`
	ExtraPerms   []Permission `json:"extra_permissions"`
	RevokedPerms []Permission `json:"revoked_permissions,omitempty"`
	Description  string       `json:"description"`
}

// OrgUnit represents an organizational unit in a hierarchy.
type OrgUnit struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	ParentID string `json:"parent_id"` // Empty for root
	Name     string `json:"name"`
}

// Delegation allows a user to delegate a permission to another user.
type Delegation struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	DelegatorID string     `json:"delegator_id"`
	DelegateeID string     `json:"delegatee_id"`
	Permission  Permission `json:"permission"`
	Resource    string     `json:"resource,omitempty"` // Optional scope
	ExpiresAt   *string    `json:"expires_at,omitempty"`
}

// AdvancedRBAC extends RBACEngine with role templates, org hierarchy, and delegation.
type AdvancedRBAC struct {
	*RBACEngine
	roleTemplates map[string]*RoleTemplate
}

// NewAdvancedRBAC creates an advanced RBAC engine.
func NewAdvancedRBAC() *AdvancedRBAC {
	return &AdvancedRBAC{
		RBACEngine:    NewRBACEngine(),
		roleTemplates: make(map[string]*RoleTemplate),
	}
}

// RegisterRoleTemplate registers a custom role template.
func (a *AdvancedRBAC) RegisterRoleTemplate(tmpl *RoleTemplate) {
	a.roleTemplates[tmpl.Name] = tmpl
}

// GetEffectiveRole returns the effective role (base + template overrides).
func (a *AdvancedRBAC) GetEffectiveRole(roleName string) (Role, []Permission) {
	if tmpl, ok := a.roleTemplates[roleName]; ok {
		basePerms := a.rolePermissions[tmpl.BaseRole]
		permSet := make(map[Permission]bool)
		for _, p := range basePerms {
			permSet[p] = true
		}
		for _, p := range tmpl.RevokedPerms {
			delete(permSet, p)
		}
		for _, p := range tmpl.ExtraPerms {
			permSet[p] = true
		}
		perms := make([]Permission, 0, len(permSet))
		for p := range permSet {
			perms = append(perms, p)
		}
		return tmpl.BaseRole, perms
	}
	if r := Role(roleName); IsValidRole(roleName) {
		return r, a.rolePermissions[r]
	}
	return RoleMember, a.rolePermissions[RoleMember]
}

// HasPermissionWithTemplate checks permission including role template overrides.
func (a *AdvancedRBAC) HasPermissionWithTemplate(roleName string, permission Permission) bool {
	_, perms := a.GetEffectiveRole(roleName)
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

// DelegationStore is the interface for storing/querying delegations.
type DelegationStore interface {
	GetActiveDelegation(ctx context.Context, tenantID, delegateeID string, permission Permission) (*Delegation, error)
}

// ValidateAccessWithDelegation validates access including delegation.
func (a *AdvancedRBAC) ValidateAccessWithDelegation(ctx context.Context, userID, tenantID string, permission Permission, store DelegationStore) error {
	// Check direct role first
	if err := a.ValidateAccess(ctx, userID, tenantID, permission); err == nil {
		return nil
	}
	// Check delegation
	if store != nil {
		del, err := store.GetActiveDelegation(ctx, tenantID, userID, permission)
		if err == nil && del != nil {
			return nil
		}
	}
	return ErrPermissionDenied
}
