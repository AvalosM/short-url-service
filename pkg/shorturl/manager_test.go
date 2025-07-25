package shorturl_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/AvalosM/meli-interview-challenge/pkg/shorturl"
	"github.com/AvalosM/meli-interview-challenge/pkg/shorturl/mocks"
)

//go:generate mockgen -typed -package=mocks  -source=./manager.go -destination=./mocks/mocks.go

type ManagerSuite struct {
	suite.Suite
	mockCtrl    *gomock.Controller
	mockStorage *mocks.MockStorage
	mockCache   *mocks.MockCache
	mockLogger  *mocks.MockLogger
	config      *shorturl.Config
	manager     *shorturl.Manager
}

func (suite *ManagerSuite) SetupTest() {
	suite.mockCtrl = gomock.NewController(suite.T())
	suite.mockStorage = mocks.NewMockStorage(suite.mockCtrl)
	suite.mockCache = mocks.NewMockCache(suite.mockCtrl)
	suite.mockLogger = mocks.NewMockLogger(suite.mockCtrl)

	suite.mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	suite.mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	suite.mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	suite.mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	suite.config = &shorturl.Config{
		MaxShortURLIdRetries:      3,
		ShortURLCacheTTLInSeconds: 60,
	}

	manager, err := shorturl.NewManager(suite.config, suite.mockStorage, suite.mockCache, suite.mockLogger)
	suite.Require().NoError(err)

	suite.manager = manager
}

func (suite *ManagerSuite) TearDownTest() {
	suite.mockCtrl.Finish()
}

func TestManagerSuite(t *testing.T) {
	suite.Run(t, new(ManagerSuite))
}

func (suite *ManagerSuite) TestGetLongURLSuccess() {
	ctx := context.Background()
	id := "AABBCC"

	expectedLongURL := "https://example.com"

	done := make(chan struct{})

	suite.mockCache.EXPECT().Get(ctx, id).Return("", false, nil)
	suite.mockStorage.EXPECT().GetLongURL(ctx, id).Return(expectedLongURL, true, nil)
	suite.mockCache.EXPECT().Set(context.WithoutCancel(ctx), id, expectedLongURL, time.Second*time.Duration(suite.config.ShortURLCacheTTLInSeconds)).
		DoAndReturn(func(ctx context.Context, s string, s2 string, duration time.Duration) error {
			close(done)
			return nil
		})

	result, err := suite.manager.GetLongURL(ctx, id)
	suite.Require().NoError(err)
	suite.Equal(expectedLongURL, result)

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		suite.Fail("Waiting for cache set timed out")
	}
}

func (suite *ManagerSuite) TestGetLongURLSuccessCacheHit() {
	ctx := context.Background()
	id := "AABBCC"

	expectedLongURL := "https://example.com"

	suite.mockCache.EXPECT().Get(ctx, id).Return(expectedLongURL, true, nil)

	result, err := suite.manager.GetLongURL(ctx, id)
	suite.Require().NoError(err)
	suite.Equal(expectedLongURL, result)
}

func (suite *ManagerSuite) TestGetLongURLFailNotFound() {
	ctx := context.Background()
	id := "AABBCC"

	suite.mockCache.EXPECT().Get(ctx, id).Return("", false, nil)
	suite.mockStorage.EXPECT().GetLongURL(ctx, id).Return("", false, nil)

	result, err := suite.manager.GetLongURL(ctx, id)
	suite.Require().ErrorIs(err, shorturl.ErrShortURLNotFound)
	suite.Zero(result)
}

func (suite *ManagerSuite) TestGetLongURLFailStorageGetLongURLError() {
	ctx := context.Background()
	id := "AABBCC"

	expectedError := errors.New("some storage error")

	suite.mockCache.EXPECT().Get(ctx, id).Return("", false, nil)
	suite.mockStorage.EXPECT().GetLongURL(ctx, id).Return("", false, expectedError)

	result, err := suite.manager.GetLongURL(ctx, id)
	suite.Require().ErrorIs(err, expectedError)
	suite.Zero(result)
}

func (suite *ManagerSuite) TestCreateShortURLSuccess() {
	ctx := context.Background()
	longURL := "https://example.com"

	expectedId, err := suite.manager.GenerateIdWithOffset(longURL, 0)
	suite.Require().NoError(err)

	suite.mockStorage.EXPECT().GetLongURL(ctx, expectedId).Return("", false, nil)
	suite.mockStorage.EXPECT().CreateShortURL(ctx, expectedId, longURL).Return(nil)

	shortURLId, err := suite.manager.CreateShortURL(ctx, longURL)
	suite.Require().NoError(err)
	suite.Equal(expectedId, shortURLId)
}

func (suite *ManagerSuite) TestCreateShortURLSuccessAlreadyExists() {
	ctx := context.Background()
	longURL := "https://example.com"

	expectedId, err := suite.manager.GenerateIdWithOffset(longURL, 0)
	suite.Require().NoError(err)

	suite.mockStorage.EXPECT().GetLongURL(ctx, expectedId).Return(longURL, true, nil)

	shortURLId, err := suite.manager.CreateShortURL(ctx, longURL)
	suite.Require().NoError(err)
	suite.Equal(expectedId, shortURLId)
}

func (suite *ManagerSuite) TestCreateShortURLSuccessHashCollision() {
	ctx := context.Background()
	longURL := "https://example.com"
	someOtherLongURL := "https://another-example.com"

	expectedId0, err := suite.manager.GenerateIdWithOffset(longURL, 0)
	suite.Require().NoError(err)

	expectedId1, err := suite.manager.GenerateIdWithOffset(longURL, 1)
	suite.Require().NoError(err)

	suite.mockStorage.EXPECT().GetLongURL(ctx, expectedId0).Return(someOtherLongURL, true, nil)
	suite.mockStorage.EXPECT().GetLongURL(ctx, expectedId1).Return("", false, nil)
	suite.mockStorage.EXPECT().CreateShortURL(ctx, expectedId1, longURL).Return(nil)

	shortURLId, err := suite.manager.CreateShortURL(ctx, longURL)
	suite.Require().NoError(err)
	suite.Equal(expectedId1, shortURLId)
}

func (suite *ManagerSuite) TestCreateShortURLFailInvalidURL() {
	ctx := context.Background()
	testCases := []struct {
		name       string
		invalidURL string
	}{
		{
			name:       "invalid URL prefix",
			invalidURL: "htp://example.com",
		},
		{
			name:       "empty URL",
			invalidURL: "",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			shortURLId, err := suite.manager.CreateShortURL(ctx, tc.invalidURL)
			suite.Require().ErrorIs(err, shorturl.ErrInvalidLongURL)
			suite.Zero(shortURLId)
		})

	}
}

func (suite *ManagerSuite) TestCreateShortURlFailMaxHashCollisions() {
	ctx := context.Background()
	longURL := "https://example.com"
	someOtherLongURL := "https://another-example.com"

	for i := range suite.config.MaxShortURLIdRetries {
		expectedId, err := suite.manager.GenerateIdWithOffset(longURL, uint(i))
		suite.Require().NoError(err)

		suite.mockStorage.EXPECT().GetLongURL(ctx, expectedId).Return(someOtherLongURL, true, nil)
	}

	shortURLId, err := suite.manager.CreateShortURL(ctx, longURL)
	suite.Require().ErrorContains(err, "failed to generate unique short URL")
	suite.Zero(shortURLId)
}

func (suite *ManagerSuite) TestCreateShortURLFailStorageGetLongURLError() {
	ctx := context.Background()
	longURL := "https://example.com"

	expectedError := errors.New("some storage error")
	expectedId, err := suite.manager.GenerateIdWithOffset(longURL, 0)
	suite.Require().NoError(err)

	suite.mockStorage.EXPECT().GetLongURL(ctx, expectedId).Return("", false, expectedError)

	shortURLId, err := suite.manager.CreateShortURL(ctx, longURL)
	suite.Require().ErrorIs(err, expectedError)
	suite.Zero(shortURLId)
}

func (suite *ManagerSuite) TestCreateShortURLFailStorageCreateShortURLError() {
	ctx := context.Background()
	longURL := "https://example.com"

	expectedError := errors.New("some storage error")
	expectedId, err := suite.manager.GenerateIdWithOffset(longURL, 0)
	suite.Require().NoError(err)

	suite.mockStorage.EXPECT().GetLongURL(ctx, expectedId).Return("", false, nil)
	suite.mockStorage.EXPECT().CreateShortURL(ctx, expectedId, longURL).Return(expectedError)

	shortURLId, err := suite.manager.CreateShortURL(ctx, longURL)
	suite.Require().ErrorIs(err, expectedError)
	suite.Zero(shortURLId)
}

func (suite *ManagerSuite) TestDeleteShortURLSuccess() {
	ctx := context.Background()
	id := "AABBCC"

	suite.mockStorage.EXPECT().DeleteShortURL(ctx, id).Return(nil)
	suite.mockCache.EXPECT().Delete(ctx, id).Return(nil)

	err := suite.manager.DeleteShortURL(ctx, id)
	suite.Require().NoError(err)
}

func (suite *ManagerSuite) TestDeleteShortURLFailStorageDeleteShortURLError() {
	ctx := context.Background()
	id := "AABBCC"

	expectedError := errors.New("some storage error")
	suite.mockStorage.EXPECT().DeleteShortURL(ctx, id).Return(expectedError)

	err := suite.manager.DeleteShortURL(ctx, id)
	suite.Require().ErrorIs(err, expectedError)
}

func (suite *ManagerSuite) TestDeleteShortURLFailCacheDeleteError() {
	ctx := context.Background()
	id := "AABBCC"

	expectedError := errors.New("some cache error")
	suite.mockStorage.EXPECT().DeleteShortURL(ctx, id).Return(nil)
	suite.mockCache.EXPECT().Delete(ctx, id).Return(expectedError)

	err := suite.manager.DeleteShortURL(ctx, id)
	suite.Require().ErrorIs(err, expectedError)
}
