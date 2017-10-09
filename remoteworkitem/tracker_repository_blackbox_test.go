package remoteworkitem_test

import (
	"testing"

	"context"

	"github.com/fabric8-services/fabric8-wit/gormsupport/cleaner"
	"github.com/fabric8-services/fabric8-wit/gormtestsupport"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type trackerRepoBlackBoxTest struct {
	gormtestsupport.DBTestSuite
	repo TrackerRepository
}

func TestRunTrackerRepoBlackBoxTest(t *testing.T) {
	suite.Run(t, &trackerRepoBlackBoxTest{DBTestSuite: gormtestsupport.NewDBTestSuite("../config.yaml")})
}

func (s *trackerRepoBlackBoxTest) SetupTest() {
	s.DBTestSuite.SetupTest()
	s.repo = NewTrackerRepository(s.DB)
}

func (s *trackerRepoBlackBoxTest) TestFailDeleteZeroID() {
	defer cleaner.DeleteCreatedEntities(s.DB)()

	// Create at least 1 item to avoid RowsEffectedCheck
	_, err := s.repo.Create(
		context.Background(),
		"http://api.github.com",
		ProviderGithub)

	if err != nil {
		s.T().Error("Could not create tracker", err)
	}

	err = s.repo.Delete(context.Background(), "0")
	require.IsType(s.T(), NotFoundError{}, err)
}

func (s *trackerRepoBlackBoxTest) TestFailSaveZeroID() {
	defer cleaner.DeleteCreatedEntities(s.DB)()

	// Create at least 1 item to avoid RowsEffectedCheck
	tr, err := s.repo.Create(
		context.Background(),
		"http://api.github.com",
		ProviderGithub)

	if err != nil {
		s.T().Error("Could not create tracker", err)
	}
	tr.ID = "0"

	_, err = s.repo.Save(context.Background(), *tr)
	require.IsType(s.T(), NotFoundError{}, err)
}

func (s *trackerRepoBlackBoxTest) TestFaiLoadZeroID() {
	defer cleaner.DeleteCreatedEntities(s.DB)()

	// Create at least 1 item to avoid RowsEffectedCheck
	_, err := s.repo.Create(
		context.Background(),
		"http://api.github.com",
		ProviderGithub)

	if err != nil {
		s.T().Error("Could not create tracker", err)
	}

	_, err = s.repo.Load(context.Background(), "0")
	var errorType NotFoundError
	require.IsType(s.T(), errorType, err)
}
