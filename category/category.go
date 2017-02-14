package category

import (
	"context"
	"time"

	"github.com/almighty/almighty-core/errors"
	"github.com/almighty/almighty-core/gormsupport"
	"github.com/goadesign/goa"
	"github.com/jinzhu/gorm"
	errs "github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// Defines "type" string to be used while validating jsonapi spec based payload
const (
	APIStringTypeCategory = "categories"
)

// Category describes a single category
type Category struct {
	gormsupport.Lifecycle
	ID          uuid.UUID `sql:"type:uuid default uuid_generate_v4()" gorm:"primary_key"`
	SpaceID     uuid.UUID `sql:"type:uuid"`
	Name        string
	Description string
}

// TableName overrides the table name settings in Gorm to force a specific table name
// in the database.
func (m *GormCategoryRepository) TableName() string {
	return "categories"
}

// NewCategoryRepository creates a new category repo
func NewCategoryRepository(db *gorm.DB) Repository {
	return &GormCategoryRepository{db: db}
}

// GormCategoryRepository implements CategoryRepository using gorm
type GormCategoryRepository struct {
	db *gorm.DB
}

// Repository describes interactions with category
type Repository interface {
	Create(ctx context.Context, u *Category) error
	List(ctx context.Context, spaceID uuid.UUID) ([]*Category, error)
	Load(ctx context.Context, id uuid.UUID) (*Category, error)
	Save(ctx context.Context, c Category) (*Category, error)
}

// Create creates a new category.
func (m *GormCategoryRepository) Create(ctx context.Context, category *Category) error {
	defer goa.MeasureSince([]string{"goa", "db", "category", "create"}, time.Now())
	category.ID = uuid.NewV4()
	err := m.db.Create(category).Error
	if err != nil {
		goa.LogError(ctx, "error adding Category", "error", err.Error())
		return errs.WithStack(err)
	}

	return nil
}

// List all Categories related to a single item
func (m *GormCategoryRepository) List(ctx context.Context, spaceID uuid.UUID) ([]*Category, error) {
	defer goa.MeasureSince([]string{"goa", "db", "category", "query"}, time.Now())
	var objs []*Category

	err := m.db.Where("space_id = ?", spaceID).Find(&objs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errs.WithStack(err)
	}
	return objs, nil
}

// Load a single Category regardless of parent
func (m *GormCategoryRepository) Load(ctx context.Context, id uuid.UUID) (*Category, error) {
	defer goa.MeasureSince([]string{"goa", "db", "category", "get"}, time.Now())
	var obj Category

	tx := m.db.Where("id = ?", id).First(&obj)
	if tx.RecordNotFound() {
		return nil, errors.NewNotFoundError("Category", id.String())
	}
	if tx.Error != nil {
		return nil, errors.NewInternalError(tx.Error.Error())
	}
	return &obj, nil
}

// Save updates the given category in the db. Version must be the same as the one in the stored version
// returns NotFoundError, VersionConflictError or InternalError
func (m *GormCategoryRepository) Save(ctx context.Context, c Category) (*Category, error) {
	cat := Category{}
	tx := m.db.Where("id=?", c.ID).First(&cat)
	if tx.RecordNotFound() {
		// treating this as a not found error: the fact that we're using number internal is implementation detail
		return nil, errors.NewNotFoundError("category", c.ID.String())
	}
	if err := tx.Error; err != nil {
		return nil, errors.NewInternalError(err.Error())
	}
	tx = tx.Save(&c)
	if err := tx.Error; err != nil {
		return nil, errors.NewInternalError(err.Error())
	}
	return &c, nil
}
