package execution

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Request struct {
	Context    context.Context
	JobID      int64
	RunID      string
	InstanceID string
	ScopeType  string
	ScopeID    *int64
	Target     *string
	OnResult   func(Result)
}

type Result struct {
	Protocol      string
	SiteID        int64
	Target        string
	Status        string
	ObservationID *int64
	DurationMS    int64
	ErrorKind     string
	ErrorMessage  string
	StartedAt     time.Time
	EndedAt       time.Time
}

var active sync.Map

// With installs lineage and scope for one protocol lane. PostgreSQL lane
// ownership guarantees that the same protocol cannot execute concurrently.
func With(protocol string, request Request, run func()) error {
	protocol = strings.TrimSpace(protocol)
	if protocol == "" || request.RunID == "" || request.JobID <= 0 {
		return fmt.Errorf("invalid durable Nav execution context")
	}
	if _, loaded := active.LoadOrStore(protocol, request); loaded {
		return fmt.Errorf("Nav protocol lane %s already has an active execution", protocol)
	}
	defer active.Delete(protocol)
	run()
	return nil
}

func Current(protocol string) (Request, bool) {
	value, ok := active.Load(protocol)
	if !ok {
		return Request{}, false
	}
	request, ok := value.(Request)
	return request, ok
}

func Allows(protocol string, siteID int64, target string) bool {
	request, ok := Current(protocol)
	if ok && request.Context != nil && request.Context.Err() != nil {
		return false
	}
	if !ok || request.ScopeType == "all" {
		return true
	}
	if request.ScopeID == nil || siteID != *request.ScopeID {
		return false
	}
	if request.ScopeType == "site" {
		return true
	}
	return request.ScopeType == "target" && request.Target != nil &&
		strings.EqualFold(strings.TrimSpace(target), strings.TrimSpace(*request.Target))
}

func Canceled(protocol string) bool {
	request, ok := Current(protocol)
	return ok && request.Context != nil && request.Context.Err() != nil
}

func TargetKey(siteID int64, target string) string {
	return strconv.FormatInt(siteID, 10) + "\x00" + strings.ToLower(strings.TrimSpace(target))
}

func Record(result Result) {
	request, ok := Current(result.Protocol)
	if !ok || request.OnResult == nil {
		return
	}
	request.OnResult(result)
}
