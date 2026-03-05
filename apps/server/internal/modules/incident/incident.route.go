package incident

import (
	"peekaping/internal/modules/middleware"

	"github.com/gin-gonic/gin"
)

type Route struct {
	controller *Controller
	middleware *middleware.AuthChain
}

func NewRoute(controller *Controller, middleware *middleware.AuthChain) *Route {
	return &Route{controller: controller, middleware: middleware}
}

func (r *Route) ConnectRoute(rg *gin.RouterGroup, controller *Controller) {
	// public route — uses /slug/:slug/ prefix to match the existing public API pattern
	sp := rg.Group("status-pages")
	sp.GET("/slug/:slug/incidents", r.controller.FindByStatusPageSlug)

	// auth-protected routes
	incidents := rg.Group("incidents")
	incidents.Use(r.middleware.AllAuth())
	incidents.POST("", r.controller.Create)
	incidents.GET("", r.controller.FindAll)
	incidents.GET("/:id", r.controller.FindByID)
	incidents.PATCH("/:id", r.controller.Update)
	incidents.PATCH("/:id/resolve", r.controller.Resolve)
	incidents.DELETE("/:id", r.controller.Delete)
}
