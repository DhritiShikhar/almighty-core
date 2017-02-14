package main

import (
	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/application"
	"github.com/almighty/almighty-core/category"
	"github.com/almighty/almighty-core/errors"
	"github.com/almighty/almighty-core/jsonapi"
	"github.com/almighty/almighty-core/login"
	"github.com/almighty/almighty-core/rest"
	"github.com/goadesign/goa"
	uuid "github.com/satori/go.uuid"
)

// SpaceCategoriesController implements the space-categories resource.
type SpaceCategoriesController struct {
	*goa.Controller
	db application.DB
}

// NewSpaceCategoriesController creates a space-categories controller.
func NewSpaceCategoriesController(service *goa.Service, db application.DB) *SpaceCategoriesController {
	return &SpaceCategoriesController{Controller: service.NewController("SpaceCategoriesController"), db: db}
}

// Create runs the create action.
func (c *SpaceCategoriesController) Create(ctx *app.CreateSpaceCategoriesContext) error {
	_, err := login.ContextIdentity(ctx)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, goa.ErrUnauthorized(err.Error()))
	}

	spaceID, err := uuid.FromString(ctx.ID)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, goa.ErrNotFound(err.Error()))
	}

	// Validate Request
	if ctx.Payload.Data == nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewBadParameterError("data", nil).Expected("not nil"))
	}
	reqCategory := ctx.Payload.Data
	if reqCategory.Attributes.Name == nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewBadParameterError("data.attributes.name", nil).Expected("not nil"))
	}

	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, goa.ErrNotFound(err.Error()))
	}

	return application.Transactional(c.db, func(appl application.Application) error {
		_, err = appl.Spaces().Load(ctx, spaceID)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, goa.ErrNotFound(err.Error()))
		}

		newCategory := category.Category{
			SpaceID: spaceID,
			Name:    *reqCategory.Attributes.Name,
		}

		if reqCategory.Attributes.Description != nil {
			newCategory.Description = *reqCategory.Attributes.Description
		}

		err = appl.Categories().Create(ctx, &newCategory)

		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}

		res := &app.CategorySingle{
			Data: ConvertCategory(ctx.RequestData, &newCategory),
		}
		ctx.ResponseData.Header().Set("Location", rest.AbsoluteURL(ctx.RequestData, app.CategoryHref(res.Data.ID)))
		return ctx.Created(res)
	})
}

// List runs the create action.
func (c *SpaceCategoriesController) List(ctx *app.ListSpaceCategoriesContext) error {
	spaceID, err := uuid.FromString(ctx.ID)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, goa.ErrNotFound(err.Error()))
	}

	return application.Transactional(c.db, func(appl application.Application) error {

		_, err = appl.Spaces().Load(ctx, spaceID)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, goa.ErrNotFound(err.Error()))
		}

		categories, err := appl.Categories().List(ctx, spaceID)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}

		res := &app.CategoryList{}
		res.Data = ConvertCategories(ctx.RequestData, categories)

		return ctx.OK(res)
	})
}
