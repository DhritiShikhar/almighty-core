package iteration

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/fabric8-services/fabric8-wit/application/repository"
	"github.com/fabric8-services/fabric8-wit/errors"
	"github.com/fabric8-services/fabric8-wit/gormsupport"
	"github.com/fabric8-services/fabric8-wit/log"
	"github.com/fabric8-services/fabric8-wit/path"
	"github.com/goadesign/goa"
	"github.com/jinzhu/gorm"
	errs "github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// Defines "type" string to be used while validating jsonapi spec based payload
const (
	APIStringTypeIteration = "iterations"
	PathSepInService       = "/"
	PathSepInDatabase      = "."
	IterationActive        = true
	IterationNotActive     = false
)

// Iteration describes a single iteration
type Iteration struct {
	gormsupport.Lifecycle
	ID          uuid.UUID `sql:"type:uuid default uuid_generate_v4()" gorm:"primary_key"` // This is the ID PK field
	SpaceID     uuid.UUID `sql:"type:uuid"`
	Path        path.Path
	StartAt     *time.Time
	EndAt       *time.Time
	Name        string
	Description *string
	State       State // this tells if iteration is currently running or not
	UserActive  bool
	// optional, private timestamp of the latest addition/removal of a relationship with this iteration
	// this field is used to generate the `ETag` and `Last-Modified` values in the HTTP responses and conditional requests processing
	RelationShipsChangedAt *time.Time `sql:"column:relationships_changed_at"`
}

// MakeChildOf does all the path magic to make the current iteration a child of
// the given parent iteration.
func (m *Iteration) MakeChildOf(parent Iteration) {
	if m.ID == uuid.Nil {
		m.ID = uuid.NewV4()
	}
	m.Path = append(parent.Path, m.ID)
}

// FullPath returns the Path by appending self ID to it.
func (m *Iteration) FullPath() string {
	return m.Path.String() + path.SepInService + m.ID.String()
}

// GetETagData returns the field values to use to generate the ETag
func (m Iteration) GetETagData() []interface{} {
	// using the 'ID' and 'UpdatedAt' (converted to number of seconds since epoch) fields
	values := []interface{}{m.ID, strconv.FormatInt(m.UpdatedAt.Unix(), 10), m.RelationShipsChangedAt}
	return values
}

// GetLastModified returns the last modification time
func (m Iteration) GetLastModified() time.Time {
	lastModified := m.UpdatedAt // default value
	// also check the optional 'RelationShipsChangedAt' field
	if m.RelationShipsChangedAt != nil && m.RelationShipsChangedAt.After(lastModified) {
		lastModified = *m.RelationShipsChangedAt
	}
	return lastModified

}

// IsRoot Checks if given iteration is a root iteration or not
func (m Iteration) IsRoot(spaceID uuid.UUID) bool {
	return m.SpaceID == spaceID && len(m.Path) == 1 && m.Path[0] == m.ID
}

// Parent returns UUID of parent iteration or uuid.Nil
// handle root itearion case, leaf node case, intermediate case
func (m Iteration) Parent() uuid.UUID {
	return m.Path.ParentID()
}

// TableName overrides the table name settings in Gorm to force a specific table name
// in the database.
func (m Iteration) TableName() string {
	return "iterations"
}

// Repository describes interactions with Iterations
type Repository interface {
	repository.Exister
	Create(ctx context.Context, u *Iteration) error
	List(ctx context.Context, spaceID uuid.UUID) ([]Iteration, error)
	Root(ctx context.Context, spaceID uuid.UUID) (*Iteration, error)
	Load(ctx context.Context, id uuid.UUID) (*Iteration, error)
	Save(ctx context.Context, i Iteration) (*Iteration, error)
	CanStart(ctx context.Context, i *Iteration) (bool, error)
	LoadMultiple(ctx context.Context, ids []uuid.UUID) ([]Iteration, error)
	LoadChildren(ctx context.Context, parentIterationID uuid.UUID) ([]Iteration, error)
	Delete(ctx context.Context, ID uuid.UUID) error
}

// NewIterationRepository creates a new storage type.
func NewIterationRepository(db *gorm.DB) Repository {
	return &GormIterationRepository{db: db}
}

// GormIterationRepository is the implementation of the storage interface for Iterations.
type GormIterationRepository struct {
	db *gorm.DB
}

// LoadMultiple returns multiple instances of iteration.Iteration
func (m *GormIterationRepository) LoadMultiple(ctx context.Context, ids []uuid.UUID) ([]Iteration, error) {
	defer goa.MeasureSince([]string{"goa", "db", "iteration", "getmultiple"}, time.Now())
	var objs []Iteration
	if len(ids) == 0 {
		return objs, nil
	}
	for i := 0; i < len(ids); i++ {
		m.db = m.db.Or("id = ?", ids[i])
	}
	tx := m.db.Find(&objs)
	if tx.Error != nil {
		return nil, errors.NewInternalError(ctx, tx.Error)
	}
	return objs, nil
}

// Create creates a new record.
func (m *GormIterationRepository) Create(ctx context.Context, u *Iteration) error {
	defer goa.MeasureSince([]string{"goa", "db", "iteration", "create"}, time.Now())

	if u.ID == uuid.Nil {
		u.ID = uuid.NewV4()
		u.Path = path.Path{u.ID}
	}
	if u.Path == nil || len(u.Path) == 0 {
		u.Path = path.Path{u.ID}
	} else {
		if u.Path.This() != u.ID {
			return errs.Errorf("iteration path has to end with iteration ID %s", u.ID)
		}
	}

	if !u.State.IsSet() {
		u.State = StateNew
	}
	err := m.db.Create(u).Error
	// Composite key (name,space,path) must be unique
	// ( name, spaceID ,path ) needs to be unique
	if gormsupport.IsUniqueViolation(err, "iterations_name_space_id_path_unique") {
		log.Error(ctx, map[string]interface{}{
			"err":      err,
			"name":     u.Name,
			"path":     u.Path,
			"space_id": u.SpaceID,
		}, "unable to create child iteration because an iteration in the same path already exists")
		return errors.NewDataConflictError(fmt.Sprintf("iteration already exists with name = %s , space_id = %s , path = %s ", u.Name, u.SpaceID.String(), u.Path.ParentPath().String()))
	}

	if err != nil {
		log.Error(ctx, map[string]interface{}{
			"iteration_id": u.ID,
			"err":          err,
		}, "unable to create the iteration")
		return errs.WithStack(err)
	}

	return nil
}

// List all Iterations related to a single item
func (m *GormIterationRepository) List(ctx context.Context, spaceID uuid.UUID) ([]Iteration, error) {
	defer goa.MeasureSince([]string{"goa", "db", "iteration", "query"}, time.Now())
	var objs []Iteration

	err := m.db.Where("space_id = ?", spaceID).Find(&objs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(ctx, map[string]interface{}{
			"space_id": spaceID,
			"err":      err,
		}, "unable to list the iterations")
		return nil, errs.WithStack(err)
	}
	return objs, nil
}

// Root returns the Root Iteration for a space
func (m *GormIterationRepository) Root(ctx context.Context, spaceID uuid.UUID) (*Iteration, error) {
	defer goa.MeasureSince([]string{"goa", "db", "iteration", "query"}, time.Now())
	var itr Iteration

	tx := m.db.Where("space_id = ? AND nlevel(path)=1", spaceID).First(&itr)
	if tx.Error != nil {
		log.Error(ctx, map[string]interface{}{
			"space_id": spaceID,
			"err":      tx.Error,
		}, "unable to get the root iteration")
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("root iteration for space", spaceID.String())
		}
		return nil, errors.NewInternalError(ctx, tx.Error)
	}
	return &itr, nil
}

// Load a single Iteration regardless of parent
func (m *GormIterationRepository) Load(ctx context.Context, id uuid.UUID) (*Iteration, error) {
	defer goa.MeasureSince([]string{"goa", "db", "iteration", "get"}, time.Now())
	var obj Iteration

	tx := m.db.Where("id = ?", id).First(&obj)
	if tx.RecordNotFound() {
		log.Error(ctx, map[string]interface{}{
			"iteration_id": id.String(),
		}, "iteration cannot be found")
		return nil, errors.NewNotFoundError("Iteration", id.String())
	}
	if tx.Error != nil {
		log.Error(ctx, map[string]interface{}{
			"iteration_id": id.String(),
			"err":          tx.Error,
		}, "unable to load the iteration")
		return nil, errors.NewInternalError(ctx, tx.Error)
	}
	return &obj, nil
}

// CheckExists returns nil if the given ID exists otherwise returns an error
func (m *GormIterationRepository) CheckExists(ctx context.Context, id uuid.UUID) error {
	defer goa.MeasureSince([]string{"goa", "db", "iteration", "exists"}, time.Now())
	return repository.CheckExists(ctx, m.db, Iteration{}.TableName(), id)
}

// Save updates the given iteration in the db. Version must be the same as the one in the stored version
// returns NotFoundError, VersionConflictError or InternalError
func (m *GormIterationRepository) Save(ctx context.Context, i Iteration) (*Iteration, error) {
	itr := Iteration{}
	tx := m.db.Where("id=?", i.ID).First(&itr)
	if tx.RecordNotFound() {
		log.Error(ctx, map[string]interface{}{
			"iteration_id": i.ID,
		}, "iteration cannot be found")
		// treating this as a not found error: the fact that we're using number internal is implementation detail
		return nil, errors.NewNotFoundError("iteration", i.ID.String())
	}
	if err := tx.Error; err != nil {
		log.Error(ctx, map[string]interface{}{
			"iteration_id": i.ID,
			"err":          err,
		}, "unknown error happened when searching the iteration")
		return nil, errors.NewInternalError(ctx, err)
	}
	tx = tx.Save(&i)
	if err := tx.Error; err != nil {
		log.Error(ctx, map[string]interface{}{
			"iteration_id": i.ID,
			"err":          err,
		}, "unable to save the iterations")
		return nil, errors.NewInternalError(ctx, err)
	}
	return &i, nil
}

// CanStart checks the rule -
// 1. Only one iteration from a space can have state=start at a time.
// 2. Root iteration of the space can not be started.(Hence can not be closed - via UI)
// Currently there is no State-machine for state transitions of iteraitons
// till then we will not allow user to start root iteration.
// More rules can be added as needed in this function
func (m *GormIterationRepository) CanStart(ctx context.Context, i *Iteration) (bool, error) {
	var count int64
	rootItr, err := m.Root(ctx, i.SpaceID)
	if err != nil {
		return false, err
	}
	if i.ID == rootItr.ID {
		return false, errors.NewBadParameterError("iteration", "Root iteration can not be started.")
	}
	m.db.Model(&Iteration{}).Where("space_id=? and state=?", i.SpaceID, StateStart).Count(&count)
	if count != 0 {
		log.Error(ctx, map[string]interface{}{
			"iteration_id": i.ID,
			"space_id":     i.SpaceID,
		}, "one iteration from given space is already running!")
		return false, errors.NewBadParameterError("state", "One iteration from given space is already running")
	}
	return true, nil
}

func inTimeframe(startAt time.Time, endAt time.Time) bool {
	return time.Now().UTC().After(startAt) && time.Now().UTC().Before(endAt)
}

func (m *Iteration) IsActive() bool {
	if m.UserActive {
		return true
	}

	if m.StartAt == nil {
		return false
	}
	if m.EndAt == nil {
		return time.Now().UTC().After(*m.StartAt)
	}
	return inTimeframe(*m.StartAt, *m.EndAt)
}

// LoadChildren executes - select * from iterations where path <@ 'parent_path.parent_id';
func (m *GormIterationRepository) LoadChildren(ctx context.Context, parentIterationID uuid.UUID) ([]Iteration, error) {
	defer goa.MeasureSince([]string{"goa", "db", "iteration", "loadchildren"}, time.Now())
	parentIteration, err := m.Load(ctx, parentIterationID)
	if err != nil {
		return nil, errors.NewNotFoundError("iteration", parentIterationID.String())
	}
	var objs []Iteration

	err = m.db.Where(fmt.Sprintf("path ~ '%s.*{1,}'", parentIteration.Path.Convert())).Order("updated_at").Find(&objs).Error
	if err != nil {
		return nil, err
	}
	return objs, nil
}

// Delete deletes the itertion with the given id
// returns NotFoundError or InternalError
func (m *GormIterationRepository) Delete(ctx context.Context, ID uuid.UUID) error {
	defer goa.MeasureSince([]string{"goa", "db", "iteration", "delete"}, time.Now())
	if ID == uuid.Nil {
		log.Error(ctx, map[string]interface{}{
			"iteration_id": ID.String(),
		}, "unable to find the iteration by ID")
		return errors.NewNotFoundError("iteration", ID.String())
	}
	itr := Iteration{ID: ID}
	tx := m.db.Delete(itr)

	if err := tx.Error; err != nil {
		log.Error(ctx, map[string]interface{}{
			"iteration_id": ID.String(),
		}, "unable to delete the iteration")
		return errors.NewInternalError(ctx, err)
	}
	if tx.RowsAffected == 0 {
		log.Error(ctx, map[string]interface{}{
			"iteration_id": ID.String(),
		}, "none row was affected by the deletion operation")
		return errors.NewNotFoundError("iteration", ID.String())
	}
	return nil
}
