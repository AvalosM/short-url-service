package storage_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/AvalosM/meli-interview-challenge/internal/storage"
	"github.com/AvalosM/meli-interview-challenge/pkg/metrics"
)

type StorageSuite struct {
	suite.Suite
	storage      *storage.Storage
	testDBName   string
	testDBConfig *storage.Config
	truncateDB   func()
}

func (suite *StorageSuite) SetupSuite() {
	suite.testDBName = "short_url_test_db"
	suite.testDBConfig = &storage.Config{
		Driver:         "pgx",
		DataSourceName: "postgres://postgres:postgres@localhost:5432",
	}

	suite.setupTestDB()

	suite.testDBConfig.DataSourceName = suite.testDBConfig.DataSourceName + fmt.Sprintf("/%s", suite.testDBName)
	storage, err := storage.NewStorage(suite.testDBConfig)
	suite.Require().NoError(err)
	suite.storage = storage
}

func (suite *StorageSuite) setupTestDB() {
	db, err := sql.Open(suite.testDBConfig.Driver, suite.testDBConfig.DataSourceName)
	suite.Require().NoError(err)

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", suite.testDBName))
	suite.Require().NoError(err)

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", suite.testDBName))
	suite.Require().NoError(err)

	err = db.Close()
	suite.Require().NoError(err)

	db, err = sql.Open(suite.testDBConfig.Driver, suite.testDBConfig.DataSourceName+"/"+suite.testDBName)
	suite.Require().NoError(err)

	query, err := os.ReadFile("../../migrations/create_tables.up.sql")
	suite.Require().NoError(err)

	_, err = db.Exec(string(query))
	suite.Require().NoError(err)

	suite.truncateDB = func() {
		_, err := db.Exec("TRUNCATE TABLE short_urls CASCADE")
		suite.Require().NoError(err)
	}
}

func (suite *StorageSuite) TearDownTest() {
	suite.truncateDB()
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}

func (suite *StorageSuite) TestCreateShortURL() {
	shortURL, longURL := "aabbcc", "https://example.com"

	err := suite.storage.CreateShortURL(context.Background(), shortURL, longURL)
	suite.Require().NoError(err)

	url, found, err := suite.storage.GetLongURL(context.Background(), shortURL)
	suite.Require().NoError(err)
	suite.True(found)
	suite.Equal(longURL, url)
}

func (suite *StorageSuite) TestDeleteShortURl() {
	shortURL, longURL := "aabbcc", "https://example.com"

	err := suite.storage.CreateShortURL(context.Background(), shortURL, longURL)
	suite.Require().NoError(err)

	err = suite.storage.DeleteShortURL(context.Background(), shortURL)
	suite.Require().NoError(err)

	_, found, err := suite.storage.GetLongURL(context.Background(), shortURL)
	suite.Require().NoError(err)
	suite.False(found)
}

func (suite *StorageSuite) TestGetLongURL() {
	shortURL, longURL := "aabbcc", "https://example.com"

	err := suite.storage.CreateShortURL(context.Background(), shortURL, longURL)
	suite.Require().NoError(err)

	url, found, err := suite.storage.GetLongURL(context.Background(), shortURL)
	suite.Require().NoError(err)
	suite.True(found)
	suite.Equal(longURL, url)
}

func (suite *StorageSuite) TestGetLongURLNotFound() {
	shortURL := "nonexistent"

	url, found, err := suite.storage.GetLongURL(context.Background(), shortURL)
	suite.Require().NoError(err)
	suite.False(found)
	suite.Empty(url)
}

func (suite *StorageSuite) TestCreateMetrics() {
	ctx := context.Background()
	shortURLId0 := "AABBCC"
	shortURLId1 := "DDEEFF"
	host0 := "127.0.0.1"
	host1 := "127.0.0.2"
	collectors := map[string]*metrics.Collector{
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

	err := suite.storage.CreateShortURL(context.Background(), shortURLId0, "https://example.com")
	suite.Require().NoError(err)

	err = suite.storage.CreateShortURL(context.Background(), shortURLId1, "https://example.com")
	suite.Require().NoError(err)

	err = suite.storage.CreateMetrics(context.Background(), collectors)
	suite.Require().NoError(err)

	retrievedMetrics, found, err := suite.storage.GetMetrics(ctx, shortURLId0, time.Now().AddDate(0, 0, -1), time.Now())
	suite.Require().NoError(err)
	suite.True(found)
	suite.Equal(collectors[shortURLId0].Visits, retrievedMetrics.Visits)
	suite.Equal(collectors[shortURLId0].UniqueVisits(), retrievedMetrics.UniqueVisits)
}

func (suite *StorageSuite) TestGetMetricsNotFound() {
	ctx := context.Background()
	shortURLId := "ababab"
	startTime := time.Now().AddDate(0, 0, -1)
	endTime := time.Now()

	_, found, err := suite.storage.GetMetrics(ctx, shortURLId, startTime, endTime)
	suite.Require().NoError(err)
	suite.False(found)
}
