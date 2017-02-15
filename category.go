package main

import (
	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/application"
	"github.com/almighty/almighty-core/category"
	"github.com/almighty/almighty-core/jsonapi"
	"github.com/almighty/almighty-core/login"
	"github.com/almighty/almighty-core/rest"
	"github.com/goadesign/goa"
	uuid "github.com/satori/go.uuid"
)

// CategoryController implements the category resource.
type CategoryController struct {
	*goa.Controller
	db application.DB
}

// NewCategoryController creates a category controller.
func NewCategoryController(service *goa.Service, db application.DB) *CategoryController {
	return &CategoryController{
		Controller: service.NewController("CategoryController"),
		db:         db,
	}
}

// CategoryConvertFunc is a open ended function to add additional links/data/relations to a Category during conversion from internal to API
type CategoryConvertFunc func(*goa.RequestData, *category.Category, *app.Category)

// ConvertCategories converts between internal and external REST representation
func ConvertCategories(request *goa.RequestData, Categories []*category.Category, additional ...CategoryConvertFunc) []*app.Category {
	var cs = []*app.Category{}
	for _, c := range Categories {
		cs = append(cs, ConvertCategory(request, c, additional...))
	}
	return cs
}

// ConvertCategory converts between internal and external REST representation
func ConvertCategory(request *goa.RequestData, c *category.Category, additional ...CategoryConvertFunc) *app.Category {
	categoryType := category.APIStringTypeCategory
	spaceType := "spaces"

	spaceID := c.SpaceID.String()

	selfURL := rest.AbsoluteURL(request, app.CategoryHref(c.ID))
	spaceSelfURL := rest.AbsoluteURL(request, app.SpaceHref(spaceID))
	workitemtypesRelatedURL := rest.AbsoluteURL(request, app.WorkitemtypeHref("?filter[category]="+c.ID.String()))

	c1 := &app.Category{
		Type: categoryType,
		ID:   &c.ID,
		Attributes: &app.CategoryAttributes{
			Name:        &c.Name,
			Description: c.Description,
		},
		Relationships: &app.CategoryRelations{
			Space: &app.RelationGeneric{
				Data: &app.GenericData{
					Type: &spaceType,
					ID:   &spaceID,
				},
				Links: &app.GenericLinks{
					Self: &spaceSelfURL,
				},
			},
			Workitemstypes: &app.RelationGeneric{
				Links: &app.GenericLinks{
					Related: &workitemtypesRelatedURL,
				},
			},
		},
		Links: &app.GenericLinks{
			Self: &selfURL,
		},
	}
	for _, add := range additional {
		add(request, c, c1)
	}
	return c1
}

// Show runs the show action.
func (c *CategoryController) Show(ctx *app.ShowCategoryContext) error {
	id, err := uuid.FromString(ctx.CategoryID)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, goa.ErrNotFound(err.Error()))
	}
	return application.Transactional(c.db, func(appl application.Application) error {
		c, err := appl.Categories().Load(ctx, id)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}

		res := &app.CategorySingle{}
		res.Data = ConvertCategory(
			ctx.RequestData,
			c)

		return ctx.OK(res)
	})
}

// Update runs the update action.
func (c *CategoryController) Update(ctx *app.UpdateCategoryContext) error {
	_, err := login.ContextIdentity(ctx)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, goa.ErrUnauthorized(err.Error()))
	}
	id, err := uuid.FromString(ctx.CategoryID)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, goa.ErrNotFound(err.Error()))
	}

	return application.Transactional(c.db, func(appl application.Application) error {
		cat, err := appl.Categories().Load(ctx.Context, id)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
		if ctx.Payload.Data.Attributes.Name != nil {
			cat.Name = *ctx.Payload.Data.Attributes.Name
		}
		if ctx.Payload.Data.Attributes.Description != nil {
			cat.Description = ctx.Payload.Data.Attributes.Description
		}
		cat, err = appl.Categories().Save(ctx.Context, *cat)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}

		response := app.CategorySingle{
			Data: ConvertCategory(ctx.RequestData, cat),
		}

		return ctx.OK(&response)
	})
}
