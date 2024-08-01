package router

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	swguicdn "github.com/swaggest/swgui/v5cdn"
	"gorm.io/gorm"

	"github.com/ingka-group-digital/ocp-go-utils/fastecho/errs"
	"github.com/ingka-group-digital/ocp-go-utils/fastecho/health"
)

// Router contains all the available routes of the service.
type Router struct {
	Routes []Route
}

// Config contains the configuration for the router.
type Config struct {
	Echo             *echo.Echo
	Routes           []Route
	SkipMetrics      bool
	SkipHealthChecks bool
	HealthChecksDB   *gorm.DB
	SwaggerTitle     string
	SwaggerPath      string
}

// Route contains the details of a route.
type Route struct {
	Group       string
	Path        string
	HandlerFunc func(ctx echo.Context) error
	RestVerb    string
}

// NewRouter creates a new Router.
func NewRouter(cfg Config) (*Router, error) {
	r := &Router{
		Routes: cfg.Routes,
	}

	if !cfg.SkipHealthChecks {
		healthHandler := health.NewHandler(cfg.HealthChecksDB)

		r.Routes = append(r.Routes, Route{
			Path:        "/health/ready",
			HandlerFunc: healthHandler.Ready,
			RestVerb:    http.MethodGet,
		})

		r.Routes = append(r.Routes, Route{
			Path:        "/health/live",
			HandlerFunc: healthHandler.Live,
			RestVerb:    http.MethodGet,
		})
	}
	err := r.setup(cfg.Echo)
	if err != nil {
		return nil, err
	}

	if !cfg.SkipMetrics {
		r.addMetrics(cfg.Echo)
	}

	r.addSwagger(cfg.Echo, cfg.SwaggerTitle, cfg.SwaggerPath)

	r.printRoutes(cfg.Echo)

	return r, nil
}

// addMetrics adds a handler for metrics.
func (r *Router) addMetrics(e *echo.Echo) *Router {
	e.GET("/metrics", echoprometheus.NewHandler())
	return r
}

const (
	swaggerPath = "/swagger"
)

// addSwagger adds a handler for swagger documentation to the given route.
func (r *Router) addSwagger(e *echo.Echo, title, path string) *Router {
	// Register the swagger.json to the server as a static resource
	e.File("swagger/swagger.json", "api/swagger.json")

	e.GET(swaggerPath, serveSwaggerUI(title, path))
	return r
}

// setup configures the routes for echo.
func (r *Router) setup(e *echo.Echo) error {
	for _, route := range r.Routes {
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

// printRoutes prints all the available routes registered in the Echo framework.
func (r *Router) printRoutes(e *echo.Echo) {
	fmt.Println("\nRegistered routes:")
	for _, route := range e.Routes() {
		fmt.Println(route.Method, " ", route.Path)
	}
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
