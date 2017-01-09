package design

import (
	d "github.com/goadesign/goa/design"
	a "github.com/goadesign/goa/design/apidsl"
)

// workItem2 defines how an update payload will look like
var workItem2 = a.Type("WorkItem2", func() {
	a.Attribute("type", d.String, func() {
		a.Enum("workitems")
	})
	a.Attribute("id", d.String, "ID of the work item which is being updated", func() {
		a.Example("42")
	})
	a.Attribute("attributes", a.HashOf(d.String, d.Any), func() {
		a.Example(map[string]interface{}{"version": "1", "system.state": "new", "system.title": "Example story"})
	})
	a.Attribute("relationships", workItemRelationships)
	a.Attribute("links", genericLinks)
	a.Required("type", "attributes")
})

// WorkItemRelationships defines only `assignee` as of now. To be updated
var workItemRelationships = a.Type("WorkItemRelationships", func() {
	a.Attribute("assignees", relationGenericList, "This defines assignees of the Work Item")
	a.Attribute("creator", relationGeneric, "This defines creator of the Work Item")
	a.Attribute("baseType", relationBaseType, "This defines type of Work Item")
	a.Attribute("comments", relationGeneric, "This defines comments on the Work Item")
	a.Attribute("iteration", relationGeneric, "This defines the iteration this work item belong to")
})

// relationBaseType is top level block for WorkItemType relationship
var relationBaseType = a.Type("RelationBaseType", func() {
	a.Attribute("data", baseTypeData)
	a.Required("data")
})

// baseTypeData is data block for `type` of a work item
var baseTypeData = a.Type("BaseTypeData", func() {
	a.Attribute("type", d.String, func() {
		a.Enum("workitemtypes")
	})
	a.Attribute("id", d.String, func() {
		a.Example("system.userstory")
	})
	a.Required("type", "id")
})

// workItemLinks has `self` as of now according to http://jsonapi.org/format/#fetching-resources
var workItemLinks = a.Type("WorkItemLinks", func() {
	a.Attribute("self", d.String, func() {
		a.Example("http://api.almighty.io/api/workitems.2/1")
	})
	a.Required("self")
})

// workItemList contains paged results for listing work items and paging links
var workItemList = JSONList(
	"WorkItem2", "Holds the paginated response to a work item list request",
	workItem2,
	meta,
	pagingLinks)

// workItemSingle is the media type for work items
var workItemSingle = JSONSingle(
	"WorkItem2", "A work item holds field values according to a given field type in JSONAPI form",
	workItem2,
	workItemLinks)

// new version of "list" for migration
var _ = a.Resource("workitem", func() {
	a.BasePath("/workitems")
	a.Action("show", func() {
		a.Routing(
			a.GET("/:id"),
		)
		a.Description("Retrieve work item with given id.")
		a.Params(func() {
			a.Param("id", d.String, "id")
		})
		a.Response(d.OK, func() {
			a.Media(workItemSingle)
		})
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
	})
	a.Action("list", func() {
		a.Routing(
			a.GET(""),
		)
		a.Description("List work items.")
		a.Params(func() {
			a.Param("filter", d.String, "a query language expression restricting the set of found work items")
			a.Param("page[offset]", d.String, "Paging start position")
			a.Param("page[limit]", d.Integer, "Paging size")
			a.Param("filter[assignee]", d.String, "Work Items assigned to the given user")
		})
		a.Response(d.OK, func() {
			a.Media(workItemList)
		})
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
	})
	a.Action("create", func() {
		a.Security("jwt")
		a.Routing(
			a.POST(""),
		)
		a.Description("create work item with type and id.")
		a.Payload(workItemSingle)
		a.Response(d.Created, "/workitems/.*", func() {
			a.Media(workItemSingle)
		})
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
	})
	a.Action("delete", func() {
		a.Security("jwt")
		a.Routing(
			a.DELETE("/:id"),
		)
		a.Description("Delete work item with given id.")
		a.Params(func() {
			a.Param("id", d.String, "id")
		})
		a.Response(d.OK)
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
	})
	a.Action("update", func() {
		a.Security("jwt")
		a.Routing(
			a.PATCH("/:id"),
		)
		a.Description("update the work item with the given id.")
		a.Params(func() {
			a.Param("id", d.String, "id")
		})
		a.Payload(workItemSingle)
		a.Response(d.OK, func() {
			a.Media(workItemSingle)
		})
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
	})
	a.Action("reorder", func() {
		a.Security("jwt")
		a.Routing(
			a.PATCH("/before/:id?reorder={id}"),
		)
		a.Description("reorder the work items before the given id.")
		a.Params(func() {
			a.Param("reorderIds", a.ArrayOf(d.String), "ids of all the workitems to be reordered")
		})
		a.Payload(workItemList)
		a.Response(d.OK, func() {
			a.Media(workItemList)
		})
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
	})
})
