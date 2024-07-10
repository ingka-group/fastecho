package router

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	swguicdn "github.com/swaggest/swgui/v5cdn"

	"github.com/ingka-group-digital/ocp-go-utils/api/core/config"
	"github.com/ingka-group-digital/ocp-go-utils/api/core/errs"
)

// Router contains all the available routes of the service.
type Router struct {
	routes []Route
}

// Route contains the details for a route.
type Route struct {
	group       *echo.Group
	path        string
	handlerFunc echo.HandlerFunc
	restVerb    string
}

// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{
		routes: []Route{},
	}
}

// AddRoute adds a route.
func (r *Router) AddRoute(group *echo.Group, path string, handlerFunc echo.HandlerFunc, restVerb string) *Router {
	r.routes = append(r.routes, Route{
		group:       group,
		path:        path,
		handlerFunc: handlerFunc,
		restVerb:    restVerb,
	})

	return r
}

// AddMetrics adds a handler for metrics e.
func (r *Router) AddMetrics(e *echo.Echo) *Router {
	e.GET("/metrics", echoprometheus.NewHandler())
	return r
}

const (
	SwaggerPath = "/swagger"
)

// AddSwagger adds a handler for swagger documentation to the given route.
func (r *Router) AddSwagger(e *echo.Echo) *Router {
	// Register the swagger.json to the server as a static resource
	e.File("swagger/swagger.json", "api/swagger.json")

	e.GET(SwaggerPath, ServeSwaggerUI())
	return r
}

// ServeSwaggerUI serves the swagger UI.
func ServeSwaggerUI() echo.HandlerFunc {
	return func(c echo.Context) error {
		swguicdn.NewHandler(
			config.Env[config.SwaggerUITitle].Value, config.Env[config.SwaggerJSONPath].Value, SwaggerPath,
		).ServeHTTP(c.Response().Writer, c.Request())

		return nil
	}
}

// Init configures the routes for echo.
func (r *Router) Init() error {
	for _, route := range r.routes {
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

// PrintRoutes prints all the available routes registered in the Echo framework.
func PrintRoutes(e *echo.Echo) {
	fmt.Println("\nRegistered routes:")
	for _, route := range e.Routes() {
		fmt.Println(route.Method, " ", route.Path)
	}
}
