package repository_test

import (
	"context"
	"testing"

	"github.com/mondegor/go-storage/mrtests/pgtest"
	"github.com/stretchr/testify/suite"

	"print-shop-back/internal/catalog/box/section/pub/entity"
	"print-shop-back/internal/catalog/box/section/pub/repository"
	"print-shop-back/tests"
)

type BoxPostgresTestSuite struct {
	suite.Suite

	ctx  context.Context
	pgt  *pgtest.Tester
	repo *repository.BoxPostgres
}

func TestBoxPostgresTestSuite(t *testing.T) {
	suite.Run(t, new(BoxPostgresTestSuite))
}

func (ts *BoxPostgresTestSuite) SetupSuite() {
	ts.ctx = context.Background()
	ts.pgt = pgtest.NewTester(ts.T(), tests.DBSchemas(), tests.ExcludedDBTables())
	ts.pgt.ApplyMigrations(ts.T(), tests.AppMigrationsDir())

	ts.repo = repository.NewBoxPostgres(ts.pgt.ConnManager())
}

func (ts *BoxPostgresTestSuite) SetupTest() {
	ts.pgt.TruncateTables(ts.T(), ts.ctx)
}

func (ts *BoxPostgresTestSuite) Test_Fetch() {
	ts.pgt.ApplyFixtures(ts.T(), "testdata/Fetch")

	expected := []entity.Box{
		{
			ID:        1,
			Article:   "T-21-310x260x380",
			Caption:   "СДЭК1",
			Length:    0.31,
			Width:     0.26,
			Height:    0.38,
			Thickness: 0.002,
			Weight:    1.05,
		},
		{
			ID:        2,
			Article:   "T-22-110x220x330",
			Caption:   "СДЭК2",
			Length:    0.11,
			Width:     0.22,
			Height:    0.33,
			Thickness: 0.0044,
			Weight:    5.55,
		},
	}

	ctx := context.Background()
	got, err := ts.repo.Fetch(ctx, entity.BoxParams{})

	ts.Require().NoError(err)
	ts.Equal(expected, got)
}
