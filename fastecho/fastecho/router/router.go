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
	Routes           func(e *echo.Echo, r *Router) error
	SkipMetrics      bool
	SkipHealthChecks bool
	HealthChecksDB   *gorm.DB
	SwaggerTitle     string
	SwaggerPath      string
}

// Route contains the details of a route.
type Route struct {
	group       *echo.Group
	path        string
	handlerFunc echo.HandlerFunc
	restVerb    string
}

// NewRouter creates a new Router.
func NewRouter(cfg Config) (*Router, error) {
	r := &Router{
		Routes: make([]Route, 0),
	}

	if !cfg.SkipHealthChecks {
		healthHandler := health.NewHandler(cfg.HealthChecksDB)

		r.Routes = append(r.Routes, Route{
			path:        "/health/ready",
			group:       cfg.Echo.Group(""),
			handlerFunc: healthHandler.Ready,
			restVerb:    http.MethodGet,
		})

		r.Routes = append(r.Routes, Route{
			path:        "/health/live",
			group:       cfg.Echo.Group(""),
			handlerFunc: healthHandler.Live,
			restVerb:    http.MethodGet,
		})
	}

	// Run the routes wrapper if it is defined.
	if cfg.Routes != nil {
		err := cfg.Routes(cfg.Echo, r)
		if err != nil {
			return nil, err
		}
	}

	err := r.setup()
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

// AddRoute adds a route.
func AddRoute(r *Router, group *echo.Group, path string, handlerFunc echo.HandlerFunc, restVerb string) *Router {
	r.Routes = append(r.Routes, Route{
		group:       group,
		path:        path,
		handlerFunc: handlerFunc,
		restVerb:    restVerb,
	})

	return r
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
func (r *Router) setup() error {
	// register routes to echo
	for _, route := range r.Routes {
		if route.group == nil {
			return errs.New("group is not defined for the route: " + route.path)
		}
		switch route.restVerb {
		case http.MethodGet:
			route.group.GET(route.path, route.handlerFunc)
		case http.MethodPost:
			route.group.POST(route.path, route.handlerFunc)
		case http.MethodPatch:
			route.group.PATCH(route.path, route.handlerFunc)
		case http.MethodDelete:
			route.group.DELETE(route.path, route.handlerFunc)
		default:
			return errs.New(
				fmt.Sprintf("not suitable router method found for: %s", route.restVerb),
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
