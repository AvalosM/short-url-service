package metrics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AvalosM/short-url-service/pkg/logging"
	"github.com/AvalosM/short-url-service/pkg/metrics"
	"github.com/AvalosM/short-url-service/pkg/metrics/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -typed -package=mocks  -source=./manager.go -destination=./mocks/mocks.go

type ManagerSuite struct {
	suite.Suite
	mockCtrl    *gomock.Controller
	mockStorage *mocks.MockStorage
	mockLogger  *mocks.MockLogger
	config      *metrics.Config
	manager     *metrics.Manager
}

func (suite *ManagerSuite) SetupTest() {
	suite.mockCtrl = gomock.NewController(suite.T())
	suite.mockStorage = mocks.NewMockStorage(suite.mockCtrl)
	suite.mockLogger = mocks.NewMockLogger(suite.mockCtrl)

	suite.mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	suite.mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	suite.mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	suite.config = metrics.DefaultConfig()

	manager, err := metrics.NewManager(suite.config, suite.mockStorage, suite.mockLogger)
	suite.Require().NoError(err)

	suite.manager = manager
}

func (suite *ManagerSuite) TearDownTest() {
	suite.mockCtrl.Finish()
}

func TestManagerSuite(t *testing.T) {
	suite.Run(t, new(ManagerSuite))
}

func (suite *ManagerSuite) TestRecordShortURLRequestAsyncSuccess() {
	shortURLId0 := "AABBCC"
	shortURLId1 := "DDEEFF"
	host0 := "127.0.0.1"
	host1 := "127.0.0.2"

	expectedCollectors := map[string]*metrics.Collector{
		shortURLId0: {
			ShortURLId: shortURLId0,
			Visits:     3,
			Visitors: map[string]struct{}{
				host0: {},
				host1: {},
			},
		},
		shortURLId1: {
			ShortURLId: shortURLId1,
			Visits:     1,
			Visitors: map[string]struct{}{
				host1: {},
			},
		},
	}

	done := make(chan struct{})

	stopManager := suite.manager.Start()

	suite.mockStorage.EXPECT().CreateMetrics(context.Background(), gomock.Any()).
		DoAndReturn(func(_ context.Context, collectors map[string]*metrics.Collector) error {
			suite.EqualValues(expectedCollectors, collectors)

			close(done)
			return nil
		})
	suite.mockStorage.EXPECT().CreateMetrics(context.Background(), gomock.Any()).AnyTimes()

	suite.manager.RecordShortURLRequestAsync(shortURLId0, host0)
	suite.manager.RecordShortURLRequestAsync(shortURLId0, host1)
	suite.manager.RecordShortURLRequestAsync(shortURLId0, host0)
	suite.manager.RecordShortURLRequestAsync(shortURLId1, host1)

	// Wait for metrics to be sent or timeout
	select {
	case <-done:
		// Successfully processed metrics
	case <-time.After(time.Duration(suite.config.MetricsIntervalInMS*2) * time.Millisecond):
		suite.T().Fatal("Timeout waiting for metrics to be processed")
	}

	stopManager()
}

func (suite *ManagerSuite) TestRecordShortURLRequestAsyncSuccessNoCollectors() {
	done := make(chan struct{})

	stopManager := suite.manager.Start()

	suite.mockStorage.EXPECT().CreateMetrics(context.Background(), map[string]*metrics.Collector{}).
		DoAndReturn(func(_ context.Context, _ map[string]*metrics.Collector) error {
			close(done)

			return nil
		})
	suite.mockStorage.EXPECT().CreateMetrics(context.Background(), gomock.Any()).AnyTimes()

	// Wait for metrics to be sent or timeout
	select {
	case <-done:
		// Successfully processed metrics
	case <-time.After(time.Duration(suite.config.MetricsIntervalInMS*2) * time.Millisecond):
		suite.T().Fatal("Timeout waiting for metrics to be processed")
	}

	stopManager()
}

func (suite *ManagerSuite) TestRecordShortURLRequestAsyncFailStorageError() {
	shortURLId0 := "AABBCC"
	shortURLId1 := "DDEEFF"
	host0 := "127.0.0.1"
	host1 := "127.0.0.2"

	expectedCollectors := map[string]*metrics.Collector{
		shortURLId0: {
			ShortURLId: shortURLId0,
			Visits:     3,
			Visitors: map[string]struct{}{
				host0: {},
				host1: {},
			},
		},
		shortURLId1: {
			ShortURLId: shortURLId1,
			Visits:     1,
			Visitors: map[string]struct{}{
				host1: {},
			},
		},
	}
	expectedError := errors.New("some storage error")

	done := make(chan struct{})

	stopManager := suite.manager.Start()

	suite.mockStorage.EXPECT().CreateMetrics(context.Background(), gomock.Any()).
		DoAndReturn(func(_ context.Context, collectors map[string]*metrics.Collector) error {
			suite.EqualValues(expectedCollectors, collectors)

			close(done)
			return expectedError
		})
	suite.mockLogger.EXPECT().Error("creating metrics in storage", logging.ErrorKey, expectedError)
	suite.mockStorage.EXPECT().CreateMetrics(context.Background(), gomock.Any()).AnyTimes()

	suite.manager.RecordShortURLRequestAsync(shortURLId0, host0)
	suite.manager.RecordShortURLRequestAsync(shortURLId0, host1)
	suite.manager.RecordShortURLRequestAsync(shortURLId0, host0)
	suite.manager.RecordShortURLRequestAsync(shortURLId1, host1)

	select {
	case <-done:
	case <-time.After(time.Duration(suite.config.MetricsIntervalInMS*2) * time.Millisecond):
		suite.Fail("Timeout waiting for metrics to be processed")
	}

	stopManager()
}

func (suite *ManagerSuite) TestGetShortURLMetricsSuccess() {
	ctx := context.Background()
	shortURLId := "AABBCC"
	from := time.Now().AddDate(0, 0, -1)
	to := time.Now()

	expectedMetrics := &metrics.Metrics{
		ShortURLId:   shortURLId,
		Visits:       42,
		UniqueVisits: 7,
		From:         from,
		To:           to,
	}

	suite.mockStorage.EXPECT().GetMetrics(ctx, shortURLId, from, to).Return(expectedMetrics, true, nil)

	metricsResult, err := suite.manager.GetShortURLMetrics(ctx, shortURLId, from, to)
	suite.Require().NoError(err)
	suite.Equal(expectedMetrics, metricsResult)
}

func (suite *ManagerSuite) TestGetShortURLMetricsSuccessMetricsNotFound() {
	ctx := context.Background()
	shortURLId := "AABBCC"
	from := time.Now().AddDate(0, 0, -1)
	to := time.Now()

	expectedMetrics := &metrics.Metrics{
		ShortURLId:   shortURLId,
		Visits:       0,
		UniqueVisits: 0,
		From:         from,
		To:           to,
	}

	suite.mockStorage.EXPECT().GetMetrics(ctx, shortURLId, from, to).Return(nil, false, nil)

	metricsResult, err := suite.manager.GetShortURLMetrics(ctx, shortURLId, from, to)
	suite.Require().NoError(err)
	suite.Equal(expectedMetrics, metricsResult)
}
