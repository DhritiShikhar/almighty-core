package main_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	. "github.com/almighty/almighty-core"
	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/app/test"
	"github.com/almighty/almighty-core/application"
	"github.com/almighty/almighty-core/category"
	"github.com/almighty/almighty-core/gormapplication"
	"github.com/almighty/almighty-core/gormsupport"
	"github.com/almighty/almighty-core/gormsupport/cleaner"
	"github.com/almighty/almighty-core/resource"
	"github.com/almighty/almighty-core/space"
	testsupport "github.com/almighty/almighty-core/test"
	almtoken "github.com/almighty/almighty-core/token"
	"github.com/goadesign/goa"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSpaceCategoryREST struct {
	gormsupport.DBTestSuite

	db    *gormapplication.GormDB
	clean func()
}

func TestRunSpaceCategoryREST(t *testing.T) {
	suite.Run(t, &TestSpaceCategoryREST{DBTestSuite: gormsupport.NewDBTestSuite("config.yaml")})
}

func (rest *TestSpaceCategoryREST) SetupTest() {
	rest.db = gormapplication.NewGormDB(rest.DB)
	rest.clean = cleaner.DeleteCreatedEntities(rest.DB)
}

func (rest *TestSpaceCategoryREST) TearDownTest() {
	rest.clean()
}

func (rest *TestSpaceCategoryREST) SecuredController() (*goa.Service, *SpaceCategoriesController) {
	priv, _ := almtoken.ParsePrivateKey([]byte(almtoken.RSAPrivateKey))

	svc := testsupport.ServiceAsUser("Category-Service", almtoken.NewManagerWithPrivateKey(priv), testsupport.TestIdentity)
	return svc, NewSpaceCategoriesController(svc, rest.db)
}

func (rest *TestSpaceCategoryREST) UnSecuredController() (*goa.Service, *SpaceCategoriesController) {
	svc := goa.New("Category-Service")
	return svc, NewSpaceCategoriesController(svc, rest.db)
}

func (rest *TestSpaceCategoryREST) TestSuccessCreateCategory() {
	t := rest.T()
	resource.Require(t, resource.Database)

	var p *space.Space
	ci := createSpaceCategory("Hello", nil)
	application.Transactional(rest.db, func(app application.Application) error {
		repo := app.Spaces()
		newSpace := space.Space{
			Name: "Test 1",
		}
		p, _ = repo.Create(context.Background(), &newSpace)
		return nil
	})
	svc, ctrl := rest.SecuredController()
	_, c := test.CreateSpaceCategoriesCreated(t, svc.Context, svc, ctrl, p.ID.String(), ci)
	require.NotNil(t, c.Data.ID)
	require.NotNil(t, c.Data.Relationships.Space)
	assert.Equal(t, p.ID.String(), *c.Data.Relationships.Space.Data.ID)
}

func (rest *TestSpaceCategoryREST) TestSuccessCreateCategoryWithOptionalValues() {
	t := rest.T()
	resource.Require(t, resource.Database)

	var p *space.Space
	categoryName := "Backlog"
	categoryDesc := "backlog workitem types"
	cat := createSpaceCategory(categoryName, &categoryDesc)
	application.Transactional(rest.db, func(app application.Application) error {
		repo := app.Spaces()
		testSpace := space.Space{
			Name: "Test 1",
		}
		p, _ = repo.Create(context.Background(), &testSpace)
		return nil
	})
	svc, ctrl := rest.SecuredController()
	_, c := test.CreateSpaceCategoriesCreated(t, svc.Context, svc, ctrl, p.ID.String(), cat)
	assert.NotNil(t, c.Data.ID)
	assert.NotNil(t, c.Data.Relationships.Space)
	assert.Equal(t, p.ID.String(), *c.Data.Relationships.Space.Data.ID)
	assert.Equal(t, *c.Data.Attributes.Name, categoryName)
	assert.Equal(t, *c.Data.Attributes.Description, categoryDesc)

	// create another Category with nil description
	categoryName2 := "Category2"
	cat = createSpaceCategory(categoryName2, nil)
	_, c = test.CreateSpaceCategoriesCreated(t, svc.Context, svc, ctrl, p.ID.String(), cat)
	assert.Equal(t, *c.Data.Attributes.Name, categoryName2)
	assert.Nil(t, c.Data.Attributes.Description)
}

func (rest *TestSpaceCategoryREST) TestListCategoriesBySpace() {
	t := rest.T()
	resource.Require(t, resource.Database)

	var spaceID uuid.UUID
	application.Transactional(rest.db, func(app application.Application) error {
		repo := app.Categories()

		newSpace := space.Space{
			Name: "Test 1",
		}
		p, err := app.Spaces().Create(context.Background(), &newSpace)
		if err != nil {
			t.Error(err)
		}
		spaceID = p.ID

		for i := 0; i < 3; i++ {
			name := "Backlog" + strconv.Itoa(i)
			c := category.Category{
				Name:    name,
				SpaceID: spaceID,
			}
			repo.Create(context.Background(), &c)
		}
		return nil
	})

	svc, ctrl := rest.UnSecuredController()
	_, cs := test.ListSpaceCategoriesOK(t, svc.Context, svc, ctrl, spaceID.String())
	assert.Len(t, cs.Data, 3)
	for _, categoryItem := range cs.Data {
		subString := fmt.Sprintf("?filter[category]=%s", categoryItem.ID.String())
		assert.Contains(t, *categoryItem.Relationships.Workitemstypes.Links.Related, subString)
	}
}

func (rest *TestSpaceCategoryREST) TestCreateCategoryMissingSpace() {
	t := rest.T()
	resource.Require(t, resource.Database)

	ci := createSpaceCategory("Backlog", nil)

	svc, ctrl := rest.SecuredController()
	test.CreateSpaceCategoriesNotFound(t, svc.Context, svc, ctrl, uuid.NewV4().String(), ci)
}

func (rest *TestSpaceCategoryREST) TestFailCreateCategoryNotAuthorized() {
	t := rest.T()
	resource.Require(t, resource.Database)

	ci := createSpaceCategory("Backlog", nil)

	svc, ctrl := rest.UnSecuredController()
	test.CreateSpaceCategoriesUnauthorized(t, svc.Context, svc, ctrl, uuid.NewV4().String(), ci)
}

func (rest *TestSpaceCategoryREST) TestFailListCategoriesByMissingSpace() {
	t := rest.T()
	resource.Require(t, resource.Database)

	svc, ctrl := rest.UnSecuredController()
	test.ListSpaceCategoriesNotFound(t, svc.Context, svc, ctrl, uuid.NewV4().String())
}

func createSpaceCategory(name string, desc *string) *app.CreateSpaceCategoriesPayload {
	return &app.CreateSpaceCategoriesPayload{
		Data: &app.Category{
			Type: category.APIStringTypeCategory,
			Attributes: &app.CategoryAttributes{
				Name:        &name,
				Description: desc,
			},
		},
	}
}
