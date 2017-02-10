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

// SpaceCategoryController implements the space-category resource.
type SpaceCategoryController struct {
	*goa.Controller
	db application.DB
}

// NewSpaceCategoryController creates a category controller.
func NewSpaceCategoryController(service *goa.Service, db application.DB) *SpaceCategoryController {
	return &SpaceCategoryController{
		Controller: service.NewController("SpaceCategoryController"),
		db:         db,
	}
}

// Create runs the create action.
func (c *SpaceCategoryController) Create(ctx *app.CreateSpaceCategoryContext) error {
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

		res = &app.CategorySingle{}
		ctx.ResponseData.Header().Set("Location", rest.AbsoluteURL(ctx.RequestData, app.CategoryHref(res.Data.ID)))
		return ctx.Created(res)
	})
}

/* Delete runs the delete action.
func (c *SpaceCategoryController) Delete(ctx *app.DeleteSpaceCategoryContext) error {
	// SpaceCategoryController_Delete: start_implement

	// Put your logic here

	// SpaceCategoryController_Delete: end_implement
	res := &app.CategorySingle{}
	return ctx.OK(res)
}

// List runs the list action.
func (c *SpaceCategoryController) List(ctx *app.ListSpaceCategoryContext) error {
	// SpaceCategoryController_List: start_implement

	// Put your logic here

	// SpaceCategoryController_List: end_implement
	res := &app.CategoryList{}
	return ctx.OK(res)
}

// Show runs the show action.
func (c *SpaceCategoryController) Show(ctx *app.ShowSpaceCategoryContext) error {
	// SpaceCategoryController_Show: start_implement

	// Put your logic here

	// SpaceCategoryController_Show: end_implement
	res := &app.CategorySingle{}
	return ctx.OK(res)
}

// Update runs the update action.
func (c *SpaceCategoryController) Update(ctx *app.UpdateSpaceCategoryContext) error {
	// SpaceCategoryController_Update: start_implement

	// Put your logic here

	// SpaceCategoryController_Update: end_implement
	res := &app.CategorySingle{}
	return ctx.OK(res)
}*/
