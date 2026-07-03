// Copyright © 2024 Ingka Holding B.V. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
