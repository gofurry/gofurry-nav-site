package authorization

import "testing"

func TestRoleCapabilityPolicy(t *testing.T) {
	tests := []struct {
		name    string
		role    Role
		allowed []Capability
		denied  []Capability
	}{
		{name: "owner", role: RoleOwner, allowed: AllCapabilities()},
		{name: "developer", role: RoleDeveloper, allowed: []Capability{
			ContentRead, ContentWrite, CollectionRead, CollectionExecute, CollectionControl,
			MetricsRead, MetricsTechnical, ChangesRead, ChangesTechnical, DataOpsRead, AuditRead,
		}, denied: []Capability{AccountManage, SystemManage}},
		{name: "operator", role: RoleOperator, allowed: []Capability{
			ContentRead, ContentWrite, CollectionRead, CollectionExecute, MetricsRead, ChangesRead,
		}, denied: []Capability{
			CollectionControl, MetricsTechnical, ChangesTechnical, DataOpsRead, AuditRead, AccountManage, SystemManage,
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, capability := range test.allowed {
				if !Has(test.role, capability) {
					t.Fatalf("role %q should have %q", test.role, capability)
				}
			}
			for _, capability := range test.denied {
				if Has(test.role, capability) {
					t.Fatalf("role %q must not have %q", test.role, capability)
				}
			}
		})
	}
}

func TestPolicyFailsClosed(t *testing.T) {
	if capabilities := CapabilitiesFor(Role("unknown")); len(capabilities) != 0 {
		t.Fatalf("unknown role received capabilities: %v", capabilities)
	}
	if Has(Role("unknown"), ContentRead) {
		t.Fatal("unknown role received content.read")
	}
	if Has(RoleOwner, Capability("unknown.capability")) {
		t.Fatal("owner received unknown capability")
	}
}

func TestOwnerHasEveryDefinedCapability(t *testing.T) {
	for _, capability := range AllCapabilities() {
		if !Has(RoleOwner, capability) {
			t.Fatalf("owner is missing %q", capability)
		}
	}
}
