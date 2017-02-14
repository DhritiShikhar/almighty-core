package design

import (
	d "github.com/goadesign/goa/design"
	a "github.com/goadesign/goa/design/apidsl"
)

var category = a.Type("Category", func() {
	a.Description("JSONAPI store for the data of a category.")
	a.Attribute("type", d.String, func() {
		a.Enum("category")
	})
	a.Attribute("id", d.UUID, "ID of category", func() {
		a.Example("40bbdd3d-8b5d-4fd6-ac90-7236b669af04")
	})
	a.Attribute("attributes", categoryAttributes)
	a.Attribute("relationships", categoryRelationships)
	a.Attribute("links", genericLinks)
	a.Required("type", "attributes")
})

var categoryAttributes = a.Type("CategoryAttributes", func() {
	a.Description(`JSONAPI store for all the "attributes" of a category. +See also see http://jsonapi.org/format/#document-resource-object-attributes`)
	a.Attribute("name", d.String, "The category name", func() {
		a.Example("Category1")
	})
	a.Attribute("description", d.String, "Description of the category", func() {
		a.Example("High level workitemtypes")
	})
})

var categoryRelationships = a.Type("CategoryRelations", func() {
	a.Attribute("space", relationGeneric, "This defines the owning space")
	a.Attribute("workitemstypes", relationGeneric, "This lines the workitemtypes associated with the category")
})

var categoryList = JSONList(
	"Category", "Holds the list of categories",
	category,
	pagingLinks,
	meta)

var categorySingle = JSONSingle(
	"Category", "Holds the list of categories",
	category,
	nil)

var _ = a.Resource("space-categories", func() {
	a.Parent("space")
	a.Action("create", func() {
		a.Security("jwt")
		a.Routing(
			a.POST("categories"),
		)
		a.Description("Create category.")
		a.Payload(categorySingle)
		a.Response(d.Created, "/categories/.*", func() {
			a.Media(categorySingle)
		})
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
	})
})
