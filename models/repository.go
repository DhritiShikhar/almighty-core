package models

import (
	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/criteria"
	"golang.org/x/net/context"
)

// WorkItemRepository encapsulates storage & retrieval of work items
type WorkItemRepository interface {
	Load(ctx context.Context, ID string) (*app.WorkItem, error)
	Save(ctx context.Context, wi app.WorkItem) (*app.WorkItem, error)
	Delete(ctx context.Context, ID string) error
	Create(ctx context.Context, typeID string, name string, fields map[string]interface{}) (*app.WorkItem, error)
	List(ctx context.Context, criteria criteria.Expression, start *int, length *int) ([]*app.WorkItem, error)
}

// TrackerRepository encapsulate storage & retrieval of tracker configuration
type TrackerRepository interface {
	Load(ctx context.Context, ID string) (*app.Tracker, error)
	Save(ctx context.Context, t app.Tracker) (*app.Tracker, error)
	Delete(ctx context.Context, ID string) error
	Create(ctx context.Context, url string, typeID string) (*app.Tracker, error)
	List(ctx context.Context, criteria criteria.Expression, start *int, length *int) ([]*app.Tracker, error)
}

// TrackerQueryRepository encapsulate storage & retrieval of tracker queries
type TrackerQueryRepository interface {
	Create(ctx context.Context, query string, schedule string, trackerID string) (*app.TrackerQuery, error)
	Save(ctx context.Context, tq app.TrackerQuery) (*app.TrackerQuery, error)
}
