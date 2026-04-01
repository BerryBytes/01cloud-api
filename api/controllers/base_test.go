package controllers

import (
	"01cloud-api/api/middlewares"
	"01cloud-api/api/models"
	"01cloud-api/api/utils/cache"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type ATServer struct {
	MActivity models.ActivityInterface
	MUser     models.UserInterface
}

func newTestServer(t *testing.T, store *ATServer) *Server {
	server, err := NewServer(store)
	require.NoError(t, err)
	return server
}

func NewServer(store *ATServer) (*Server, error) {
	mockactivity := store
	server := &Server{
		MockInterface: mockactivity,
		Router:        mux.NewRouter(),
		Cache:         cache.NewCache(1 * time.Minute),
	}
	if aserver, ok := server.MockInterface.(*ATServer); ok {
		activityInterface = aserver.MActivity
		middlewares.UserI = aserver.MUser

	}
	server.initializeRoutes("http://api.example.com")

	return server, nil
}

func TestMain(m *testing.M) {
	var activityInterface models.ActivityInterface
	var subscriptionInterface models.SubscriptionInterface
	fmt.Println(activityInterface, subscriptionInterface)
	os.Exit(m.Run())
}
