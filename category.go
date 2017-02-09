package main

import (
	"github.com/almighty/almighty-core/application"
	"github.com/almighty/almighty-core/category"
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

/* Create runs the create action.
func (c *CategoryController) Create(ctx *app.CreateSpaceCategoryContext) error {
	_, err := login.ContextIdentity(ctx)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, goa.ErrUnauthorized(err.Error()))
	}
	err = validateCreateCategory(ctx)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, err)
	}

	return application.Transactional(c.db, func(appl application.Application) error {
		reqCategory := ctx.Payload.Data

		newCategory := category.Category{
			SpaceID:     nil,
			Name:        *reqCategory.Attributes.Name,
			Description: *reqCategory.Attributes.Description,
		}
		category, err = appl.Categories().Create(ctx, &newCategory)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}

		res := &app.CategorySingle{
			Data: ConvertCategory(ctx.RequestData, category),
		}

		res := &app.CategorySingle{}
		ctx.ResponseData.Header().Set("Location", rest.AbsoluteURL(ctx.RequestData, app.CategoryHref(res.Data.ID)))
		return ctx.Created(res)
	})
}

// Delete runs the delete action.
func (c *CategoryController) Delete(ctx *app.DeleteCategoryContext) error {
	// CategoryController_Delete: start_implement

	// Put your logic here

	// CategoryController_Delete: end_implement
	res := &app.CategorySingle{}
	return ctx.OK(res)
}

// List runs the list action.
func (c *CategoryController) List(ctx *app.ListCategoryContext) error {
	// CategoryController_List: start_implement

	// Put your logic here

	// CategoryController_List: end_implement
	res := &app.CategoryList{}
	return ctx.OK(res)
}

// Show runs the show action.
func (c *CategoryController) Show(ctx *app.ShowCategoryContext) error {
	// CategoryController_Show: start_implement

	// Put your logic here

	// CategoryController_Show: end_implement
	res := &app.CategorySingle{}
	return ctx.OK(res)
}

// Update runs the update action.
func (c *CategoryController) Update(ctx *app.UpdateCategoryContext) error {
	// CategoryController_Update: start_implement

	// Put your logic here

	// CategoryController_Update: end_implement
	res := &app.CategorySingle{}
	return ctx.OK(res)
}

func validateCreateCategory(ctx *app.CreateCategoryContext) error {
	if ctx.Payload.Data == nil {
		return errors.NewBadParameterError("data", nil).Expected("not nil")
	}
	if ctx.Payload.Data.Attributes == nil {
		return errors.NewBadParameterError("data.attributes", nil).Expected("not nil")
	}
	if ctx.Payload.Data.Name == nil {
		return errors.NewBadParameterError("data.name", nil).Expected("not nil")
	}
}*/

// ConvertCategory converts between internal and external REST representation
func ConvertCategory(request *goa.RequestData, c *category.Category, additional ...CategoryConvertFunc) *app.Category {
	categoryType := c.APIStringTypeIteration
	spaceType := "spaces"

	spaceID := c.SpaceID.String()

	selfURL := rest.AbsoluteURL(request, app.CategoryHref(c.ID))
	spaceSelfURL := rest.AbsoluteURL(request, app.SpaceHref(spaceID))
	workitemsRelatedURL := rest.AbsoluteURL(request, app.WorkitemHref("?filter[iteration]="+itr.ID.String()))

	i := &app.Iteration{
		Type: iterationType,
		ID:   &itr.ID,
		Attributes: &app.IterationAttributes{
			Name:        &itr.Name,
			StartAt:     itr.StartAt,
			EndAt:       itr.EndAt,
			Description: itr.Description,
			State:       &itr.State,
		},
		Relationships: &app.IterationRelations{
			Space: &app.RelationGeneric{
				Data: &app.GenericData{
					Type: &spaceType,
					ID:   &spaceID,
				},
				Links: &app.GenericLinks{
					Self: &spaceSelfURL,
				},
			},
			Workitems: &app.RelationGeneric{
				Links: &app.GenericLinks{
					Related: &workitemsRelatedURL,
				},
			},
		},
		Links: &app.GenericLinks{
			Self: &selfURL,
		},
	}
	if itr.ParentID != uuid.Nil {
		parentSelfURL := rest.AbsoluteURL(request, app.IterationHref(itr.ParentID))
		parentID := itr.ParentID.String()
		i.Relationships.Parent = &app.RelationGeneric{
			Data: &app.GenericData{
				Type: &iterationType,
				ID:   &parentID,
			},
			Links: &app.GenericLinks{
				Self: &parentSelfURL,
			},
		}
	}
	for _, add := range additional {
		add(request, itr, i)
	}
	return i
}
