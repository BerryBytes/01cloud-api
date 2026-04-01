package registry

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/stretchr/testify.v1/require"
)

type MockDomainClient struct {
	mock.Mock
}

func TestNewDockerHubRegistrty(t *testing.T) {
	_ = os.Setenv("DOCKERHUB_USERNAME", "")
	_ = os.Setenv("DOCKERHUB_TOKEN", "")

	_ = os.Setenv("DOCKERHUB_USERNAME", "")
	_ = os.Unsetenv("DOCKERHUB_TOKEN")
	_, err := NewDockerHubRegistrty(context.Background(), os.Getenv("DOCKERHUB_USERNAME"), os.Getenv("DOCKERHUB_TOKEN"))
	require.Error(t, err)
}

func TestDockerParseWebhook(t *testing.T) {
	docker := &DockerRegistry{}
	inputs := []DockerhubWebhook{
		{
			PushData: PushData{
				Tag: "Latest",
			},
			Repository: Repo{
				RepoURL: "http://www.test.com",
			},
		},
		{
			PushData: PushData{
				Tag: "Latest",
			},
			Repository: Repo{
				RepoURL: "http://www.test.com/hook",
			},
		},
	}

	for _, data := range inputs {
		request, err := json.Marshal(data)
		if err != nil {
			log.Error("err", err)
		}
		resp, err := docker.ParseWebhook(request)

		if err != nil {
			log.Error("err 2nd", err)
		}

		assert.Equal(t, resp.Tag, data.PushData.Tag)
		assert.Equal(t, resp.URL, data.Repository.RepoURL)
	}
}
