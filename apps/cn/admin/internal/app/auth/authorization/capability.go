package authorization

type Capability string

const (
	ContentRead       Capability = "content.read"
	ContentWrite      Capability = "content.write"
	CollectionRead    Capability = "collection.read"
	CollectionExecute Capability = "collection.execute"
	CollectionControl Capability = "collection.control"
	MetricsRead       Capability = "metrics.read"
	MetricsTechnical  Capability = "metrics.technical"
	ChangesRead       Capability = "changes.read"
	ChangesTechnical  Capability = "changes.technical"
	DataOpsRead       Capability = "dataops.read"
	AuditRead         Capability = "audit.read"
	AccountManage     Capability = "account.manage"
	SystemManage      Capability = "system.manage"
	CloudOpsRead      Capability = "cloudops.read"
	CloudOpsManage    Capability = "cloudops.manage"
	CloudOpsPurgeAll  Capability = "cloudops.purge_all"
)

var capabilityCatalog = []Capability{
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
	AccountManage,
	SystemManage,
	CloudOpsRead,
	CloudOpsManage,
	CloudOpsPurgeAll,
}

func AllCapabilities() []Capability {
	return append([]Capability(nil), capabilityCatalog...)
}

func IsDefinedCapability(capability Capability) bool {
	for _, defined := range capabilityCatalog {
		if capability == defined {
			return true
		}
	}
	return false
}
