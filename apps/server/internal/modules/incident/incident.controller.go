package incident

import (
	"net/http"
	"peekaping/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Controller struct {
	service Service
	logger  *zap.SugaredLogger
}

func NewController(service Service, logger *zap.SugaredLogger) *Controller {
	return &Controller{
		service: service,
		logger:  logger,
	}
}

// @Router		/incidents [get]
// @Summary		Get incidents
// @Tags		Incidents
// @Produce		json
// @Security	JwtAuth
// @Security	ApiKeyAuth
// @Param		q     query  string  false  "Search query"
// @Param		page  query  int     false  "Page number" default(1)
// @Param		limit query  int     false  "Items per page" default(10)
// @Success		200	{object}	utils.ApiResponse[[]Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) FindAll(ctx *gin.Context) {
	page, err := utils.GetQueryInt(ctx, "page", 0)
	if err != nil || page < 0 {
		ctx.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid page parameter"))
		return
	}

	limit, err := utils.GetQueryInt(ctx, "limit", 10)
	if err != nil || limit < 1 {
		ctx.JSON(http.StatusBadRequest, utils.NewFailResponse("Invalid limit parameter"))
		return
	}

	q := ctx.Query("q")

	entities, err := c.service.FindAll(ctx, page, limit, q)
	if err != nil {
		c.logger.Errorw("Failed to fetch incidents", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, utils.NewSuccessResponse("success", entities))
}

// @Router		/incidents [post]
// @Summary		Create incident
// @Tags		Incidents
// @Produce		json
// @Accept		json
// @Security	JwtAuth
// @Security	ApiKeyAuth
// @Param		body body CreateDto true "Incident object"
// @Success		201	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) Create(ctx *gin.Context) {
	var entity *CreateDto
	if err := ctx.ShouldBindJSON(&entity); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
		return
	}

	if err := utils.Validate.Struct(entity); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
		return
	}

	created, err := c.service.Create(ctx, entity)
	if err != nil {
		c.logger.Errorw("Failed to create incident", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	ctx.JSON(http.StatusCreated, utils.NewSuccessResponse("Incident created successfully", created))
}

// @Router		/incidents/{id} [get]
// @Summary		Get incident by ID
// @Tags		Incidents
// @Produce		json
// @Security	JwtAuth
// @Security	ApiKeyAuth
// @Param		id path string true "Incident ID"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) FindByID(ctx *gin.Context) {
	id := ctx.Param("id")

	entity, err := c.service.FindByID(ctx, id)
	if err != nil {
		c.logger.Errorw("Failed to fetch incident", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	if entity == nil {
		ctx.JSON(http.StatusNotFound, utils.NewFailResponse("Incident not found"))
		return
	}

	ctx.JSON(http.StatusOK, utils.NewSuccessResponse("success", entity))
}

// @Router		/incidents/{id} [patch]
// @Summary		Update incident
// @Tags		Incidents
// @Produce		json
// @Accept		json
// @Security	JwtAuth
// @Security	ApiKeyAuth
// @Param		id   path string    true "Incident ID"
// @Param		body body UpdateDto true "Incident update object"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		400	{object}	utils.APIError[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var entity UpdateDto
	if err := ctx.ShouldBindJSON(&entity); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
		return
	}

	if err := utils.Validate.Struct(entity); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.NewFailResponse(err.Error()))
		return
	}

	updated, err := c.service.Update(ctx, id, &entity)
	if err != nil {
		c.logger.Errorw("Failed to update incident", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	if updated == nil {
		ctx.JSON(http.StatusNotFound, utils.NewFailResponse("Incident not found"))
		return
	}

	ctx.JSON(http.StatusOK, utils.NewSuccessResponse("Incident updated successfully", updated))
}

// @Router		/incidents/{id}/resolve [patch]
// @Summary		Resolve incident
// @Tags		Incidents
// @Produce		json
// @Security	JwtAuth
// @Security	ApiKeyAuth
// @Param		id path string true "Incident ID"
// @Success		200	{object}	utils.ApiResponse[Model]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) Resolve(ctx *gin.Context) {
	id := ctx.Param("id")

	resolved, err := c.service.Resolve(ctx, id)
	if err != nil {
		c.logger.Errorw("Failed to resolve incident", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	if resolved == nil {
		ctx.JSON(http.StatusNotFound, utils.NewFailResponse("Incident not found"))
		return
	}

	ctx.JSON(http.StatusOK, utils.NewSuccessResponse("Incident resolved successfully", resolved))
}

// @Router		/incidents/{id} [delete]
// @Summary		Delete incident
// @Tags		Incidents
// @Produce		json
// @Security	JwtAuth
// @Security	ApiKeyAuth
// @Param		id path string true "Incident ID"
// @Success		200	{object}	utils.ApiResponse[any]
// @Failure		404	{object}	utils.APIError[any]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	existing, err := c.service.FindByID(ctx, id)
	if err != nil {
		c.logger.Errorw("Failed to fetch incident", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	if existing == nil {
		ctx.JSON(http.StatusNotFound, utils.NewFailResponse("Incident not found"))
		return
	}

	err = c.service.Delete(ctx, id)
	if err != nil {
		c.logger.Errorw("Failed to delete incident", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, utils.NewSuccessResponse[any]("Incident deleted successfully", nil))
}

// @Router		/status-pages/slug/{slug}/incidents [get]
// @Summary		Get incidents for a status page
// @Tags		Status Pages
// @Produce		json
// @Param		slug path string true "Status page slug"
// @Success		200	{object}	utils.ApiResponse[[]Model]
// @Failure		500	{object}	utils.APIError[any]
func (c *Controller) FindByStatusPageSlug(ctx *gin.Context) {
	slug := ctx.Param("slug")

	incidents, err := c.service.FindByStatusPageSlug(ctx, slug, 0, 100)
	if err != nil {
		c.logger.Errorw("Failed to fetch incidents by slug", "error", err)
		ctx.JSON(http.StatusInternalServerError, utils.NewFailResponse("Internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, utils.NewSuccessResponse("success", incidents))
}
