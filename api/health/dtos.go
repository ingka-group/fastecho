package health

type ServiceHealthStatus string

type ServiceHealth struct {
	ServiceStatus ServiceHealthStatus `json:"status"`
}

var (
	StatusAlive       ServiceHealthStatus = "ok"
	StatusUnavailable ServiceHealthStatus = "unavailable"
)
