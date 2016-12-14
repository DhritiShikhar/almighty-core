package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/application"
	"github.com/almighty/almighty-core/errors"
	"github.com/almighty/almighty-core/jsonapi"
	"github.com/almighty/almighty-core/login"
	query "github.com/almighty/almighty-core/query/simple"
	"github.com/almighty/almighty-core/workitem"
	"github.com/goadesign/goa"
	uuid "github.com/satori/go.uuid"
)

const (
	pageSizeDefault = 20
	pageSizeMax     = 100

	APIStringTypeUser         = "identities"
	APIStringTypeWorkItem     = "workitems"
	APIStringTypeWorkItemType = "workitemtypes"
)

// WorkitemController implements the workitem resource.
type WorkitemController struct {
	*goa.Controller
	db application.DB
}

// NewWorkitemController creates a workitem controller.
func NewWorkitemController(service *goa.Service, db application.DB) *WorkitemController {
	if db == nil {
		panic("db must not be nil")
	}
	return &WorkitemController{Controller: service.NewController("WorkitemController"), db: db}
}

func parseInts(s *string) ([]int, error) {
	if s == nil || len(*s) == 0 {
		return []int{}, nil
	}
	split := strings.Split(*s, ",")
	result := make([]int, len(split))
	for index, value := range split {
		converted, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		result[index] = converted
	}
	return result, nil
}

func parseLimit(pageParameter *string) (s *int, l int, e error) {
	params, err := parseInts(pageParameter)
	if err != nil {
		return nil, 0, err
	}

	if len(params) > 1 {
		return &params[0], params[1], nil
	}
	if len(params) > 0 {
		return nil, params[0], nil
	}
	return nil, 100, nil
}

func buildAbsoluteURL(req *goa.RequestData) string {
	scheme := "http"
	if req.TLS != nil { // isHTTPS
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s", scheme, req.Host, req.URL.Path)
}

func setPagingLinks(links *app.PagingLinks, path string, resultLen, offset, limit, count int) {

	// prev link
	if offset > 0 && count > 0 {
		var prevStart int
		// we do have a prev link
		if offset <= count {
			prevStart = offset - limit
		} else {
			// the first range that intersects the end of the useful range
			prevStart = offset - (((offset-count)/limit)+1)*limit
		}
		realLimit := limit
		if prevStart < 0 {
			// need to cut the range to start at 0
			realLimit = limit + prevStart
			prevStart = 0
		}
		prev := fmt.Sprintf("%s?page[offset]=%d&page[limit]=%d", path, prevStart, realLimit)
		links.Prev = &prev
	}

	// next link
	nextStart := offset + resultLen
	if nextStart < count {
		// we have a next link
		next := fmt.Sprintf("%s?page[offset]=%d&page[limit]=%d", path, nextStart, limit)
		links.Next = &next
	}

	// first link
	var firstEnd int
	if offset > 0 {
		firstEnd = offset % limit // this is where the second page starts
	} else {
		// offset == 0, first == current
		firstEnd = limit
	}
	first := fmt.Sprintf("%s?page[offset]=%d&page[limit]=%d", path, 0, firstEnd)
	links.First = &first

	// last link
	var lastStart int
	if offset < count {
		// advance some pages until touching the end of the range
		lastStart = offset + (((count - offset - 1) / limit) * limit)
	} else {
		// retreat at least one page until covering the range
		lastStart = offset - ((((offset - count) / limit) + 1) * limit)
	}
	realLimit := limit
	if lastStart < 0 {
		// need to cut the range to start at 0
		realLimit = limit + lastStart
		lastStart = 0
	}
	last := fmt.Sprintf("%s?page[offset]=%d&page[limit]=%d", path, lastStart, realLimit)
	links.Last = &last
}

// List runs the list action.
// Prev and Next links will be present only when there actually IS a next or previous page.
// Last will always be present. Total Item count needs to be computed from the "Last" link.
func (c *WorkitemController) List(ctx *app.ListWorkitemContext) error {
	// Workitem2Controller_List: start_implement

	exp, err := query.Parse(ctx.Filter)
	if err != nil {
		jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrBadRequest(fmt.Sprintf("could not parse filter: %s", err.Error())))
		return ctx.BadRequest(jerrors)
	}
	var offset int
	var limit int

	if ctx.PageOffset == nil {
		offset = 0
	} else {
		offsetValue, err := strconv.Atoi(*ctx.PageOffset)
		if err != nil {
			offset = 0
		} else {
			offset = offsetValue
		}
	}
	if offset < 0 {
		offset = 0
	}

	if ctx.PageLimit == nil {
		limit = pageSizeDefault
	} else {
		limit = *ctx.PageLimit
	}

	if limit <= 0 {
		limit = pageSizeDefault
	} else if limit > pageSizeMax {
		limit = pageSizeMax
	}

	return application.Transactional(c.db, func(tx application.Application) error {
		result, tc, err := tx.WorkItems().List(ctx.Context, exp, &offset, &limit)
		count := int(tc)
		if err != nil {
			switch err := err.(type) {
			case errors.BadParameterError:
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrBadRequest(fmt.Sprintf("Error listing work items: %s", err.Error())))
				return ctx.BadRequest(jerrors)
			default:
				log.Printf("Error listing work items: %s", err.Error())
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrInternal(fmt.Sprintf("Error listing work items: %s", err.Error())))
				return ctx.InternalServerError(jerrors)
			}
		}

		response := app.WorkItem2List{
			Links: &app.PagingLinks{},
			Meta:  &app.WorkItemListResponseMeta{TotalCount: count},
			Data:  ConvertWorkItems(ctx.RequestData, result),
		}

		setPagingLinks(response.Links, buildAbsoluteURL(ctx.RequestData), len(result), offset, limit, count)

		return ctx.OK(&response)
	})

	// Workitem2Controller_List: end_implement
}

// Update does PATCH workitem
func (c *WorkitemController) Update(ctx *app.UpdateWorkitemContext) error {
	return application.Transactional(c.db, func(appl application.Application) error {

		if ctx.Payload == nil || ctx.Payload.Data == nil || ctx.Payload.Data.ID == nil {
			jerrors, _ := jsonapi.ErrorToJSONAPIErrors(errors.NewBadParameterError("data.id", nil))
			return ctx.NotFound(jerrors)
		}

		wi, err := appl.WorkItems().Load(ctx, *ctx.Payload.Data.ID)
		if err != nil {
			jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrNotFound(fmt.Sprintf("Error updating work item: %s", err.Error())))
			return ctx.NotFound(jerrors)
		}
		err = c.ConvertJSONAPIToWorkItem(*ctx.Payload.Data, wi)
		if err != nil {
			jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrBadRequest(fmt.Sprintf("Error updating work item: %s", err.Error())))
			return ctx.BadRequest(jerrors)
		}
		wi, err = appl.WorkItems().Save(ctx, *wi)
		if err != nil {
			switch err := err.(type) {
			case errors.BadParameterError:
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrBadRequest(fmt.Sprintf("Error updating work item: %s", err.Error())))
				return ctx.BadRequest(jerrors)
			case errors.NotFoundError:
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrNotFound(err.Error()))
				return ctx.NotFound(jerrors)
			case errors.VersionConflictError:
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrBadRequest(fmt.Sprintf("Error updating work item: %s", err.Error())))
				return ctx.BadRequest(jerrors)
			default:
				log.Printf("Error updating work items: %s", err.Error())
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrInternal(err.Error()))
				return ctx.InternalServerError(jerrors)
			}
		}

		wi2 := ConvertWorkItem(ctx.RequestData, wi)
		resp := &app.WorkItem2Single{
			Data: wi2,
			Links: &app.WorkItemLinks{
				Self: buildAbsoluteURL(ctx.RequestData),
			},
		}

		return ctx.OK(resp)
	})
}

// Create does POST workitem
func (c *WorkitemController) Create(ctx *app.CreateWorkitemContext) error {
	currentUser, err := login.ContextIdentity(ctx)
	if err != nil {
		jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrUnauthorized(err.Error()))
		return ctx.Unauthorized(jerrors)
	}

	var wit *string
	if ctx.Payload.Data != nil && ctx.Payload.Data.Relationships != nil && ctx.Payload.Data.Relationships.BaseType != nil {
		if ctx.Payload.Data.Relationships.BaseType.Data != nil {
			wit = &ctx.Payload.Data.Relationships.BaseType.Data.ID
		}
	}
	if wit == nil { // TODO Figure out path source etc. Should be a required relation
		jerrors, _ := jsonapi.ErrorToJSONAPIErrors(errors.NewBadParameterError("data.relationships.basetype.data.id", nil))
		return ctx.BadRequest(jerrors)

	}

	wi := app.WorkItem{
		Fields: make(map[string]interface{}),
	}
	c.ConvertJSONAPIToWorkItem(*ctx.Payload.Data, &wi)

	return application.Transactional(c.db, func(appl application.Application) error {

		wi, err := appl.WorkItems().Create(ctx, *wit, wi.Fields, wi.Previtemid, currentUser)
		if err != nil {
			switch err := err.(type) {
			case errors.BadParameterError:
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrBadRequest(fmt.Sprintf("Error updating work item: %s", err.Error())))
				return ctx.BadRequest(jerrors)
			case errors.VersionConflictError:
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrBadRequest(fmt.Sprintf("Error updating work item: %s", err.Error())))
				return ctx.BadRequest(jerrors)
			default:
				log.Printf("Error updating work items: %s", err.Error())
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrInternal(err.Error()))
				return ctx.InternalServerError(jerrors)
			}
		}

		wi2 := ConvertWorkItem(ctx.RequestData, wi)
		resp := &app.WorkItem2Single{
			Data: wi2,
			Links: &app.WorkItemLinks{
				Self: buildAbsoluteURL(ctx.RequestData),
			},
		}
		ctx.ResponseData.Header().Set("Location", app.WorkitemHref(wi2.ID))
		return ctx.Created(resp)
	})
}

// Show does GET workitem
func (c *WorkitemController) Show(ctx *app.ShowWorkitemContext) error {
	return application.Transactional(c.db, func(appl application.Application) error {

		comments := WorkItemIncludeCommentsAndTotal(ctx, c.db, ctx.ID)

		wi, err := appl.WorkItems().Load(ctx, ctx.ID)
		if err != nil {
			switch err := err.(type) {
			case errors.NotFoundError:
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrNotFound(err.Error()))
				return ctx.NotFound(jerrors)
			case errors.BadParameterError:
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrBadRequest(fmt.Sprintf("Error updating work item: %s", err.Error())))
				return ctx.BadRequest(jerrors)
			case errors.VersionConflictError:
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrBadRequest(fmt.Sprintf("Error updating work item: %s", err.Error())))
				return ctx.BadRequest(jerrors)
			default:
				log.Printf("Error updating work items: %s", err.Error())
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrInternal(err.Error()))
				return ctx.InternalServerError(jerrors)
			}
		}

		wi2 := ConvertWorkItem(ctx.RequestData, wi, comments)
		resp := &app.WorkItem2Single{
			Data: wi2,
		}
		return ctx.OK(resp)
	})
}

// Delete does DELETE workitem
func (c *WorkitemController) Delete(ctx *app.DeleteWorkitemContext) error {
	return application.Transactional(c.db, func(appl application.Application) error {

		err := appl.WorkItems().Delete(ctx, ctx.ID)
		if err != nil {
			switch err := err.(type) {
			case errors.NotFoundError:
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrNotFound(err.Error()))
				return ctx.NotFound(jerrors)
			case errors.BadParameterError:
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrBadRequest(fmt.Sprintf("Error updating work item: %s", err.Error())))
				return ctx.BadRequest(jerrors)
			case errors.VersionConflictError:
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrBadRequest(fmt.Sprintf("Error updating work item: %s", err.Error())))
				return ctx.BadRequest(jerrors)
			default:
				log.Printf("Error updating work items: %s", err.Error())
				jerrors, _ := jsonapi.ErrorToJSONAPIErrors(goa.ErrInternal(err.Error()))
				return ctx.InternalServerError(jerrors)
			}
		}
		return ctx.OK([]byte{})
	})
}

// ConvertJSONAPIToWorkItem is responsible for converting given WorkItem model object into a
// response resource object by jsonapi.org specifications
func (c *WorkitemController) ConvertJSONAPIToWorkItem(source app.WorkItem2, target *app.WorkItem) error {
	// construct default values from input WI

	var version = -1
	if source.Attributes["version"] != nil {
		v, err := strconv.Atoi(fmt.Sprintf("%v", source.Attributes["version"]))
		if err != nil {
			return errors.NewBadParameterError("data.attributes.version", source.Attributes["version"])
		}
		version = v
	}
	target.Version = version

	if source.Relationships != nil && source.Relationships.Assignees != nil {
		if source.Relationships.Assignees.Data == nil {
			delete(target.Fields, workitem.SystemAssignees)
		} else {
			var ids []string
			for _, d := range source.Relationships.Assignees.Data {
				assigneeUUID, err := uuid.FromString(*d.ID)
				if err != nil {
					return errors.NewBadParameterError("data.relationships.assignees.data.id", *d.ID)
				}
				ok := c.db.Identities().ValidIdentity(context.Background(), assigneeUUID)
				if !ok {
					return errors.NewBadParameterError("data.relationships.assignees.data.id", *d.ID)
				}
				ids = append(ids, assigneeUUID.String())
			}

			target.Fields[workitem.SystemAssignees] = ids
		}
	}
	if source.Relationships != nil && source.Relationships.BaseType != nil {
		if source.Relationships.BaseType.Data != nil {
			target.Type = source.Relationships.BaseType.Data.ID
		}
	}
	for key, val := range source.Attributes {
		target.Fields[key] = val
	}
	return nil
}

// WorkItemConvertFunc is a open ended function to add additional links/data/relations to a Comment during
// convertion from internal to API
type WorkItemConvertFunc func(*goa.RequestData, *app.WorkItem, *app.WorkItem2)

// ConvertWorkItems is responsible for converting given []WorkItem model object into a
// response resource object by jsonapi.org specifications
func ConvertWorkItems(request *goa.RequestData, wis []*app.WorkItem, additional ...WorkItemConvertFunc) []*app.WorkItem2 {
	ops := []*app.WorkItem2{}
	for _, wi := range wis {
		ops = append(ops, ConvertWorkItem(request, wi, additional...))
	}
	return ops
}

// ConvertWorkItem is responsible for converting given WorkItem model object into a
// response resource object by jsonapi.org specifications
func ConvertWorkItem(request *goa.RequestData, wi *app.WorkItem, additional ...WorkItemConvertFunc) *app.WorkItem2 {
	// construct default values from input WI
	selfURL := AbsoluteURL(request, app.WorkitemHref(wi.ID))
	op := &app.WorkItem2{
		ID:   &wi.ID,
		Type: APIStringTypeWorkItem,
		Attributes: map[string]interface{}{
			"version":    wi.Version,
			"previtemid": wi.Previtemid,
		},
		Relationships: &app.WorkItemRelationships{
			BaseType: &app.RelationBaseType{
				Data: &app.BaseTypeData{
					ID:   wi.Type,
					Type: APIStringTypeWorkItemType,
				},
			},
		},
		Links: &app.GenericLinks{
			Self: &selfURL,
		},
	}

	// Move fields into Relationships or Attributes as needed
	for name, val := range wi.Fields {
		switch name {
		case workitem.SystemAssignees:
			if val != nil {
				valArr := val.([]interface{})
				op.Relationships.Assignees = &app.RelationGenericList{
					Data: ConvertUsersSimple(request, valArr),
				}
			}
		case workitem.SystemCreator:
			if val != nil {
				valStr := val.(string)
				op.Relationships.Creator = &app.RelationGeneric{
					Data: ConvertUserSimple(request, valStr),
				}
			}
		default:
			op.Attributes[name] = val
		}
	}
	if op.Relationships.Assignees == nil {
		op.Relationships.Assignees = &app.RelationGenericList{Data: nil}
	}
	// Always include Comments Link, but optionally use WorkItemIncludeCommentsAndTotal
	WorkItemIncludeComments(request, wi, op)
	for _, add := range additional {
		add(request, wi, op)
	}
	return op
}
