package router

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	swguicdn "github.com/swaggest/swgui/v5cdn"

	"github.com/ingka-group-digital/ocp-go-utils/api/core/config"
	"github.com/ingka-group-digital/ocp-go-utils/api/core/env"
	"github.com/ingka-group-digital/ocp-go-utils/api/core/errs"
	"github.com/ingka-group-digital/ocp-go-utils/api/health"
)

// Router contains all the available routes of the service.
type Router struct {
	options config.Options
	routes  []Route
}

// Route contains the details of a route.
type Route struct {
	Group       string
	Path        string
	HandlerFunc func(ctx echo.Context) error
	RestVerb    string
}

// NewRouter creates a new Router.
func NewRouter(routes []Route, opts config.Options) *Router {
	// bind routes
	r := []Route{}
	r = append(r, routes...)

	return &Router{
		routes:  r,
		options: opts,
	}
}

func (r *Router) RegisterRoutes(e *echo.Echo, envs env.EnvVars) error {
	if !r.options.HealthChecks.Skip {
		healthHandler := health.NewHealthHandler(r.options.HealthChecks.DB)
		r.routes = append(r.routes, Route{
			Path:        "/health/ready",
			HandlerFunc: healthHandler.Ready,
			RestVerb:    http.MethodGet,
		})
		r.routes = append(r.routes, Route{
			Path:        "/health/live",
			HandlerFunc: healthHandler.Live,
			RestVerb:    http.MethodGet,
		})
	}
	err := r.setup(e)
	if err != nil {
		return err
	}

	if !r.options.SkipMetrics {
		r.addMetrics(e)
	}
	if !r.options.SkipSwagger {
		r.addSwagger(e, envs[env.SwaggerUITitle].Value, envs[env.SwaggerJSONPath].Value)
	}

	printRoutes(e)

	return nil
}

// AddMetrics adds a handler for metrics e.
func (r *Router) addMetrics(e *echo.Echo) *Router {
	e.GET("/metrics", echoprometheus.NewHandler())
	return r
}

const (
	swaggerPath = "/swagger"
)

// AddSwagger adds a handler for swagger documentation to the given route.
func (r *Router) addSwagger(e *echo.Echo, title, path string) *Router {
	// Register the swagger.json to the server as a static resource
	e.File("swagger/swagger.json", "api/swagger.json")

	e.GET(swaggerPath, serveSwaggerUI(title, path))
	return r
}

// serveSwaggerUI serves the swagger UI.
func serveSwaggerUI(title, path string) echo.HandlerFunc {
	return func(c echo.Context) error {
		swguicdn.NewHandler(
			title, path, swaggerPath,
		).ServeHTTP(c.Response().Writer, c.Request())

		return nil
	}
}

// setup configures the routes for echo.
func (r *Router) setup(e *echo.Echo) error {
	for _, route := range r.routes {
		switch route.RestVerb {
		case http.MethodGet:
			e.Group(route.Group).GET(route.Path, route.HandlerFunc)
		case http.MethodPost:
			e.Group(route.Group).POST(route.Path, route.HandlerFunc)
		case http.MethodPatch:
			e.Group(route.Group).PATCH(route.Path, route.HandlerFunc)
		case http.MethodDelete:
			e.Group(route.Group).DELETE(route.Path, route.HandlerFunc)
		default:
			return errs.New(
				fmt.Sprintf("not suitable router method found for: %s", route.RestVerb),
			)
		}
	}

	return nil
}

// PrintRoutes prints all the available routes registered in the Echo framework.
func printRoutes(e *echo.Echo) {
	fmt.Println("\nRegistered routes:")
	for _, route := range e.Routes() {
		fmt.Println(route.Method, " ", route.Path)
	}
}
