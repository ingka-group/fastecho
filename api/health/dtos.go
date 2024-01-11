package health

import "time"

// ServiceHealth defines the health of the service.
type ServiceHealth struct {
	ServiceStatus ServiceHealthStatus      `json:"status"`
	Description   ServiceHealthDescription `json:"description"`
	CompletedAt   time.Time                `json:"completedAt"`
} // @name ServiceHealth

// ServiceHealthStatus defines the status of the service.
type ServiceHealthStatus string // @name ServiceHealthStatus

const (
	StatusHealthy   ServiceHealthStatus = "healthy"
	StatusUnhealthy ServiceHealthStatus = "unhealthy"
)

// ServiceHealthDescription describes the state of the service status.
type ServiceHealthDescription string // @name ServiceHealthDescription

const (
	DescriptionHealthy        ServiceHealthDescription = "everything is awesome"
	DescriptionDatabaseIsDown ServiceHealthDescription = "database is down"
)
