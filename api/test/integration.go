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

type IntegrationTest struct {
	T        *testing.T
	Db       *gorm.DB
	Echo     *echo.Echo
	Fixtures *Fixtures

	opts []IntegrationTestOption
}

// NewIntegrationTest prepares database for integration testing
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

type IntegrationTestWithDatabase struct{}

func (o IntegrationTestWithDatabase) setup(it *IntegrationTest) {
	dbContainer, err := setupPostgresDB(context.Background())
	if err != nil {
		it.T.Fatalf("database setup error: %v", err)
	}

	dsn := dbURL(dbContainer.DBHost, nat.Port(fmt.Sprintf("%d/tcp", dbContainer.DBPort)))
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		it.T.Fatalf("database connection error: %v", err)
	}

	it.Db = db
}

func (o IntegrationTestWithDatabase) tearDown(it *IntegrationTest) {
	//
}
