package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// IntegrationTest is a struct that holds all the necessary information for integration testing.
type IntegrationTest struct {
	T           *testing.T
	Db          *gorm.DB
	Echo        *echo.Echo
	Fixtures    *Fixtures
	Container   *PostgresDBContainer
	BqContainer *BigqueryEmulatorContainer
	Mock        *Mock

	opts []IntegrationTestOption
}

// NewIntegrationTest prepares database for integration testing.
func NewIntegrationTest(t *testing.T, opts ...IntegrationTestOption) *IntegrationTest {
	it := &IntegrationTest{
		T:        t,
		Echo:     echo.New(),
		Fixtures: &Fixtures{},
	}

	for _, o := range opts {
		o.setup(it)
	}

	it.opts = opts

	return it
}

func (it *IntegrationTest) TearDown() {
	for _, o := range it.opts {
		o.tearDown(it)
	}
}

type IntegrationTestOption interface {
	setup(*IntegrationTest)
	tearDown(*IntegrationTest)
}

// IntegrationTestWithPostgres is an option for integration testing that sets up a postgres database test container.
// In the InitSQLScript a SQL script filename can be passed to initialize the database. The script should be located
// under a 'fixtures' directory where the _test.go file is located. An optional gorm config can also be passed.
type IntegrationTestWithPostgres struct {
	InitSQLScript string
	Config        *gorm.Config
}

func (o IntegrationTestWithPostgres) setup(it *IntegrationTest) {
	// sanity check
	if o.Config == nil {
		o.Config = &gorm.Config{}
	}

	dbContainer, err := setupPostgresDB(context.Background(), o.InitSQLScript)
	if err != nil {
		it.T.Fatalf("database setup error: %v", err)
	}

	it.Container = dbContainer

	dsn := dbURL(dbContainer.DBHost, nat.Port(fmt.Sprintf("%d/tcp", dbContainer.DBPort)))
	db, err := gorm.Open(postgres.Open(dsn), o.Config)
	if err != nil {
		it.T.Fatalf("database connection error: %v", err)
	}

	it.Db = db
}

func (o IntegrationTestWithPostgres) tearDown(it *IntegrationTest) {
	err := it.Container.Terminate(context.Background())
	if err != nil {
		it.T.Logf("error detected during container termination: %v", err)
	}
}

// IntegrationTestWithMocks is an option for integration testing that allows mocking
// The mocks should be placed in a 'mocks' directory where the _test.go file is located.
type IntegrationTestWithMocks struct {
	BaseURL string
}

func (o IntegrationTestWithMocks) setup(it *IntegrationTest) {
	it.Mock = NewMock(o.BaseURL)
}

func (o IntegrationTestWithMocks) tearDown(it *IntegrationTest) {
	it.Mock.TearDown()
}

// IntegrationTestWithBigQuery is an option for integration testing that sets up a BigQuery database test container.
type IntegrationTestWithBigQuery struct {
	DataPath string
}

func (o IntegrationTestWithBigQuery) setup(it *IntegrationTest) {
	container, err := setupBigqueryEmulator(context.Background(), o.DataPath)
	if err != nil {
		it.T.Fatalf("database setup error: %v", err)
	}

	it.BqContainer = container
}

func (o IntegrationTestWithBigQuery) tearDown(it *IntegrationTest) {
	err := it.BqContainer.Terminate(context.Background())
	if err != nil {
		it.T.Logf("error detected during container termination: %v", err)
	}
}
