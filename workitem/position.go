package workitem

import (
	"github.com/almighty/almighty-core/convert"
	"github.com/almighty/almighty-core/gormsupport"
)

type Position struct {
	gormsupport.Lifecycle
	// Id of the work item which is to be moved
	Id uint64 `gorm:"primary_key"`
	// Id of the work item above which we are going to move the selected work item
	PrevItemId int
}

// TableName implements gorm.tabler
func (p Position) TableName() string {
	return "position"
}

// Ensure Position implements the Equaler interface
var _ convert.Equaler = Position{}
var _ convert.Equaler = (*Position)(nil)

// Equal returns true if two WorkItem objects are equal; otherwise false is returned.
func (p Position) Equal(u convert.Equaler) bool {
	other, ok := u.(Position)
	if !ok {
		return false
	}
	if !p.Lifecycle.Equal(other.Lifecycle) {
		return false
	}
	if p.Id != other.Id {
		return false
	}
	if p.PrevItemId != other.PrevItemId {
		return false
	}
	return true
}
