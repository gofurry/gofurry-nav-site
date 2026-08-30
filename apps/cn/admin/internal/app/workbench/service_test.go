package workbench

import (
	"testing"

	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
)

func TestFeaturesFollowCapabilitiesInsteadOfRoles(t *testing.T) {
	tests := []struct {
		name         string
		capabilities []authorization.Capability
		assert       func(t *testing.T, flags featureFlags)
	}{
		{name: "operator", capabilities: []authorization.Capability{authorization.CollectionRead, authorization.MetricsRead, authorization.ChangesRead}, assert: func(t *testing.T, flags featureFlags) {
			if !flags.collection || !flags.metrics || !flags.changes || flags.metricTechnical || flags.dataOps || flags.audit || flags.accounts {
				t.Fatalf("operator flags=%+v", flags)
			}
		}},
		{name: "developer", capabilities: []authorization.Capability{authorization.CollectionRead, authorization.MetricsRead, authorization.MetricsTechnical, authorization.ChangesRead, authorization.ChangesTechnical, authorization.DataOpsRead, authorization.AuditRead}, assert: func(t *testing.T, flags featureFlags) {
			if !flags.metricTechnical || !flags.changeTechnical || !flags.dataOps || !flags.audit || flags.accounts {
				t.Fatalf("developer flags=%+v", flags)
			}
		}},
		{name: "owner governance", capabilities: []authorization.Capability{authorization.AccountManage}, assert: func(t *testing.T, flags featureFlags) {
			if !flags.accounts || flags.collection {
				t.Fatalf("capability-only flags=%+v", flags)
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			principal := &authorization.Principal{Capabilities: test.capabilities}
			test.assert(t, featuresFor(principal))
		})
	}
}
