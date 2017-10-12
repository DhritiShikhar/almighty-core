package application

import (
	"github.com/fabric8-services/fabric8-wit/workitem"

	"context"
)

/* TrackerRepository encapsulate storage & retrieval of tracker configuration
type TrackerRepository interface {
	repository.Exister
	Load(ctx context.Context, ID uuid.UUID) (*remoteworkitem.Tracker, error)
	Save(ctx context.Context, t *remoteworkitem.Tracker) (*remoteworkitem.Tracker, error)
	Delete(ctx context.Context, ID uuid.UUID) error
	Create(ctx context.Context, t *remoteworkitem.Tracker) error
	List(ctx context.Context) ([]remoteworkitem.Tracker, error)
}

// TrackerQueryRepository encapsulate storage & retrieval of tracker queries
type TrackerQueryRepository interface {
	repository.Exister
	Create(ctx context.Context, query string, schedule string, tracker string, spaceID uuid.UUID) (*app.TrackerQuery, error)
	Save(ctx context.Context, tq app.TrackerQuery) (*app.TrackerQuery, error)
	Load(ctx context.Context, ID string) (*app.TrackerQuery, error)
	Delete(ctx context.Context, ID string) error
	List(ctx context.Context) ([]*app.TrackerQuery, error)
}*/

// SearchRepository encapsulates searching of woritems,users,etc
type SearchRepository interface {
	SearchFullText(ctx context.Context, searchStr string, start *int, length *int, spaceID *string) ([]workitem.WorkItem, uint64, error)
	Filter(ctx context.Context, filterStr string, parentExists *bool, start *int, length *int) ([]workitem.WorkItem, uint64, error)
}
