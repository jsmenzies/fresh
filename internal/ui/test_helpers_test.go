package ui

import (
	"fresh/internal/domain"
	"fresh/internal/testhelpers"
)

type testRepositoryBuilder = testhelpers.RepositoryBuilder

func makeTestRepository(name string) domain.Repository {
	return newTestRepository(name).Build()
}

func newTestRepository(name string) *testRepositoryBuilder {
	return testhelpers.NewTestRepository(name)
}
