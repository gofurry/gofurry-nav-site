package authorization

const PrincipalContextKey = "auth_principal"

type Principal struct {
	AccountID      int64         `json:"account_id"`
	Username       string        `json:"username"`
	DisplayName    string        `json:"display_name"`
	Role           Role          `json:"role"`
	Status         AccountStatus `json:"status"`
	SessionVersion int64         `json:"session_version"`
	Capabilities   []Capability  `json:"capabilities"`
}

func (principal Principal) Has(capability Capability) bool {
	if !IsDefinedCapability(capability) {
		return false
	}
	for _, granted := range principal.Capabilities {
		if granted == capability {
			return true
		}
	}
	return false
}

var rolePolicy = map[Role][]Capability{
	RoleOwner: AllCapabilities(),
	RoleDeveloper: {
		ContentRead,
		ContentWrite,
		CollectionRead,
		CollectionExecute,
		CollectionControl,
		MetricsRead,
		MetricsTechnical,
		ChangesRead,
		ChangesTechnical,
		DataOpsRead,
		AuditRead,
	},
	RoleOperator: {
		ContentRead,
		ContentWrite,
		CollectionRead,
		CollectionExecute,
		MetricsRead,
		ChangesRead,
	},
}

func CapabilitiesFor(role Role) []Capability {
	capabilities, ok := rolePolicy[role]
	if !ok {
		return nil
	}
	return append([]Capability(nil), capabilities...)
}

func Has(role Role, capability Capability) bool {
	if !IsDefinedCapability(capability) {
		return false
	}
	for _, granted := range rolePolicy[role] {
		if granted == capability {
			return true
		}
	}
	return false
}
