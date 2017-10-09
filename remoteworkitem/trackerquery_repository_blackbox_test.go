package remoteworkitem_test

import (
	"testing"

	"github.com/fabric8-services/fabric8-wit/gormtestsupport"
	"github.com/fabric8-services/fabric8-wit/space"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type trackerQueryRepoBlackBoxTest struct {
	gormtestsupport.DBTestSuite
	repo   TrackerQueryRepository
	trRepo TrackerRepository
}

func TestRunTrackerQueryRepoBlackBoxTest(t *testing.T) {
	suite.Run(t, &trackerQueryRepoBlackBoxTest{DBTestSuite: gormtestsupport.NewDBTestSuite("../config.yaml")})
}

func (s *trackerQueryRepoBlackBoxTest) SetupTest() {
	s.DBTestSuite.SetupTest()
	s.repo = NewTrackerQueryRepository(s.DB)
	s.trRepo = NewTrackerRepository(s.DB)
}

func (s *trackerQueryRepoBlackBoxTest) TestFailDeleteZeroID() {
	// Create at least 1 item to avoid RowsEffectedCheck
	tr, err := s.trRepo.Create(
		s.Ctx,
		"http://api.github.com",
		ProviderGithub)
	if err != nil {
		s.T().Error("Could not create tracker", err)
	}

	_, err = s.repo.Create(
		s.Ctx,
		"project = ARQ AND text ~ 'arquillian'",
		"15 * * * * *",
		tr.ID, space.SystemSpace)
	if err != nil {
		s.T().Error("Could not create tracker query", err)
	}

	err = s.repo.Delete(s.Ctx, "0")
	require.IsType(s.T(), NotFoundError{}, err)
}

func (s *trackerQueryRepoBlackBoxTest) TestFailSaveZeroID() {
	// Create at least 1 item to avoid RowsEffectedCheck
	tr, err := s.trRepo.Create(
		s.Ctx,
		"http://api.github.com",
		ProviderGithub)
	if err != nil {
		s.T().Error("Could not create tracker", err)
	}

	tq, err := s.repo.Create(
		s.Ctx,
		"project = ARQ AND text ~ 'arquillian'",
		"15 * * * * *",
		tr.ID, space.SystemSpace)
	if err != nil {
		s.T().Error("Could not create tracker query", err)
	}
	tq.ID = "0"

	_, err = s.repo.Save(s.Ctx, *tq)
	require.IsType(s.T(), NotFoundError{}, err)
}

func (s *trackerQueryRepoBlackBoxTest) TestFaiLoadZeroID() {
	// Create at least 1 item to avoid RowsEffectedCheck
	tr, err := s.trRepo.Create(
		s.Ctx,
		"http://api.github.com",
		ProviderGithub)
	if err != nil {
		s.T().Error("Could not create tracker", err)
	}

	_, err = s.repo.Create(
		s.Ctx,
		"project = ARQ AND text ~ 'arquillian'",
		"15 * * * * *",
		tr.ID, space.SystemSpace)
	if err != nil {
		s.T().Error("Could not create tracker query", err)
	}

	_, err = s.repo.Load(s.Ctx, "0")
	require.IsType(s.T(), NotFoundError{}, err)
}
