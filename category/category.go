package category

import (
	"context"
	"time"

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
	SpaceID     uuid.UUID `sql:"type:uuid`
	Name        string
	Description string
}

/* Equal returns true if two Category objects are equal; otherwise false is returned.
func (p Category) Equal(u convert.Equaler) bool {
	other, ok := u.(Category)
	if !ok {
		return false
	}
	lfEqual := p.Lifecycle.Equal(other.Lifecycle)
	if !lfEqual {
		return false
	}
	if p.Name != other.Name {
		return false
	}
	if p.Description != other.Description {
		return false
	}
	return true
}*/

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
