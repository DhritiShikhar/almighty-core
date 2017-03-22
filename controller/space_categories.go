package controller

import (
	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/application"
	"github.com/almighty/almighty-core/category"
	"github.com/almighty/almighty-core/jsonapi"
	"github.com/almighty/almighty-core/rest"
	"github.com/goadesign/goa"
	uuid "github.com/satori/go.uuid"
)

// SpaceCategoriesController implements the space_categories resource.
type SpaceCategoriesController struct {
	*goa.Controller
	db application.DB
}

// NewSpaceCategoriesController creates a space_categories controller.
func NewSpaceCategoriesController(service *goa.Service, db application.DB) *SpaceCategoriesController {
	return &SpaceCategoriesController{Controller: service.NewController("SpaceCategoriesController"), db: db}
}

// List runs the list action.
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
		categories, err := appl.Categories().List(ctx)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
		res := &app.CategoryList{}
		res.Data = ConvertCategories(ctx.RequestData, categories)

		return ctx.OK(res)
	})
}

// ConvertCategories converts between internal and external REST representation
func ConvertCategories(request *goa.RequestData, Categories []*category.Category) []*app.Categories {
	var cs = []*app.Categories{}
	for _, c := range Categories {
		cs = append(cs, ConvertCategory(request, c))
	}
	return cs
}

// ConvertCategories converts between internal and external REST representation
func ConvertCategory(request *goa.RequestData, c *category.Category) *app.Categories {
	categoryType := category.APIStringTypeCategories
	selfURL := rest.AbsoluteURL(request, app.IterationHref(c.ID))
	category := &app.Categories{
		Type: categoryType,
		ID:   &c.ID,
		Attributes: &app.CategoryAttributes{
			Name: c.Name,
		},
		Relationships: &app.CategoryRelations{},
		Links: &app.GenericLinks{
			Self: &selfURL,
		},
	}
	return category
}
