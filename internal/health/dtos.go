package health

import "time"

// ServiceHealth defines the health of the service.
type ServiceHealth struct {
	ServiceStatus ServiceHealthStatus      `json:"status"`
	Description   ServiceHealthDescription `json:"description"`
	CompletedAt   time.Time                `json:"completed_at"`
} // @name ServiceHealth

// ServiceHealthStatus defines the status of the service.
type ServiceHealthStatus string // @name serviceHealthStatus

const (
	statusHealthy   ServiceHealthStatus = "healthy"
	statusUnhealthy ServiceHealthStatus = "unhealthy"
)

// ServiceHealthDescription describes the state of the service status.
type ServiceHealthDescription string // @name serviceHealthDescription

const (
	descriptionHealthy        ServiceHealthDescription = "everything is awesome"
	descriptionDatabaseIsDown ServiceHealthDescription = "database is down"
)
