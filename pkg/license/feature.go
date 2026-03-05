// Package license provides feature gating for open-core: OSS core is MIT; paid add-ons (SSO, Compliance Pack, etc.) can be gated.
package license

// Feature is a gated capability. OSS core features are always enabled; enterprise features require a valid license.
type Feature string

const (
	// Core (always available in OSS)
	FeaturePolicyEngine    Feature = "policy_engine"
	FeatureAuditExport     Feature = "audit_export"
	FeatureWorkflowMarket  Feature = "workflow_marketplace"
	FeatureCostTracking    Feature = "cost_tracking"
	FeatureAirGappedDeploy Feature = "air_gapped"

	// Enterprise (gated; require license in commercial build)
	FeatureSSO             Feature = "sso"
	FeatureCompliancePack  Feature = "compliance_pack"
	FeatureHostedTier      Feature = "hosted_tier"
	FeaturePaidMarketplace Feature = "paid_marketplace"
)

// Checker returns whether a feature is enabled. Default implementation is OSS-only (core features on, enterprise off).
type Checker interface {
	Enabled(f Feature) bool
}

// OSSChecker is the default: all core features enabled, all enterprise features disabled.
type OSSChecker struct{}

func (OSSChecker) Enabled(f Feature) bool {
	switch f {
	case FeaturePolicyEngine, FeatureAuditExport, FeatureWorkflowMarket, FeatureCostTracking, FeatureAirGappedDeploy:
		return true
	case FeatureSSO, FeatureCompliancePack, FeatureHostedTier, FeaturePaidMarketplace:
		return false
	default:
		return false
	}
}

// DefaultChecker is the default Checker used when no license is configured (OSS mode).
var DefaultChecker Checker = OSSChecker{}

// Enabled reports whether feature f is enabled under the default checker.
func Enabled(f Feature) bool {
	return DefaultChecker.Enabled(f)
}
