package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBeforeDelete(t *testing.T) {

	Repo := &Repo{
		Name: "good",
	}
	Repo.ID = "1"
	Chart := Chart{
		Name:   "chart",
		RepoID: "1",
	}
	Chart.ID = "1"
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewRepo()
	err := data.BeforeDelete(server.DB, Repo)
	assert.Error(t, err)

}
