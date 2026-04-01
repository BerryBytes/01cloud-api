package controllers

import (
	mockdb "01cloud-api/api/controllers/mocks"
	"01cloud-api/api/middlewares"
	"01cloud-api/api/models"
	"01cloud-api/api/utils/cache"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LoadbalancerTestServer struct {
	MUser         models.UserInterface
	MProject      models.IProject
	MLoadBalancer models.LoadBalancerInterface
	MOrg          models.OrganizationInterface
}

func newLoadbalancerTestServer(t *testing.T, store *LoadbalancerTestServer) *Server {
	fmt.Println("jhj", store)
	server, err := NewLoadbalancerServer(store)
	require.NoError(t, err)
	return server
}

func NewLoadbalancerServer(store *LoadbalancerTestServer) (*Server, error) {
	server := &Server{
		MockInterface: store,
		Router:        mux.NewRouter(),
		Cache:         cache.NewCache(1 * time.Minute),
	}
	if pserver, ok := server.MockInterface.(*LoadbalancerTestServer); ok {
		fmt.Println("mmm", pserver)
		iproject = pserver.MProject
		userInterface = pserver.MUser
		orgInterface = pserver.MOrg
		loadbalancerInterface = pserver.MLoadBalancer
		middlewares.UserI = pserver.MUser
	}
	server.initializeRoutes("http://api.example.com")
	return server, nil
}
func TestGETLoadBalancerAPI(t *testing.T) {
	tProject := &models.Project{
		Name:           "test-pro",
		Description:    "testing purpose",
		ProjectCode:    "TTTTT",
		UserID:         1,
		SubscriptionID: 1,
	}
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		IsAdmin:   true,
		Company:   "test",
	}
	user.ID = 1
	tProject.ID = 1
	loadbalancer := &[]models.LoadBalancer{
		{
			Name:         "nnn",
			CustomDomain: "uuu",
			ProjectID:    1,
		},
	}
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(lstore *mockdb.MockLoadBalancerInterface, pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(lstore *mockdb.MockLoadBalancerInterface, pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {

				pstore.EXPECT().
					Find(gomock.Any(), gomock.Eq(uint64(tProject.ID))).
					Times(1).
					Return(tProject, nil)
				lstore.EXPECT().
					FindAllByProject(gomock.Any(), gomock.Eq(uint64(tProject.ID))).
					Times(1).
					Return(loadbalancer, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			pstore := mockdb.NewMockIProject(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			lstore := mockdb.NewMockLoadBalancerInterface(ctrl)
			ostore := mockdb.NewMockOrganizationInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(lstore, pstore, ustore)
			tLoadbalancerServer := &LoadbalancerTestServer{
				MUser:         ustore,
				MProject:      pstore,
				MLoadBalancer: lstore,
				MOrg:          ostore,
			}
			pserver := newLoadbalancerTestServer(t, tLoadbalancerServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/project/%v/loadbalancers", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			pserver.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}

func TestFetchLoadBalancerAPI(t *testing.T) {
	tProject := &models.Project{
		Name:           "test-pro",
		Description:    "testing purpose",
		ProjectCode:    "TTTTT",
		UserID:         1,
		SubscriptionID: 1,
	}
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		IsAdmin:   true,
		Company:   "test",
	}
	user.ID = 1
	tProject.ID = 1
	loadbalancer := &models.LoadBalancer{

		Name:         "nnn",
		CustomDomain: "uuu",
		ProjectID:    1,
	}
	loadbalancer.ID = 1
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(lstore *mockdb.MockLoadBalancerInterface, pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(lstore *mockdb.MockLoadBalancerInterface, pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
				lstore.EXPECT().
					Find(gomock.Any(), gomock.Eq(uint64(loadbalancer.ID))).
					Times(1).
					Return(&models.LoadBalancer{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "OKError",
			PID:  "rrr",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(lstore *mockdb.MockLoadBalancerInterface, pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			pstore := mockdb.NewMockIProject(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			lstore := mockdb.NewMockLoadBalancerInterface(ctrl)
			ostore := mockdb.NewMockOrganizationInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(lstore, pstore, ustore)
			tLoadbalancerServer := &LoadbalancerTestServer{
				MUser:         ustore,
				MProject:      pstore,
				MLoadBalancer: lstore,
				MOrg:          ostore,
			}
			pserver := newLoadbalancerTestServer(t, tLoadbalancerServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/loadbalancer/%v/fetch-status", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			pserver.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}

func TestGetLoadBalancerAPI(t *testing.T) {
	tProject := &models.Project{
		Name:           "test-pro",
		Description:    "testing purpose",
		ProjectCode:    "TTTTT",
		UserID:         1,
		SubscriptionID: 1,
	}
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		IsAdmin:   true,
		Company:   "test",
	}
	user.ID = 1
	tProject.ID = 1
	loadbalancer := &models.LoadBalancer{
		Name:         "nnn",
		CustomDomain: "uuu",
		ProjectID:    1,
	}
	loadbalancer.ID = 1
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(lstore *mockdb.MockLoadBalancerInterface, pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(lstore *mockdb.MockLoadBalancerInterface, pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
				lstore.EXPECT().
					Find(gomock.Any(), gomock.Eq(uint64(loadbalancer.ID))).
					Times(1).
					Return(loadbalancer, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "OKError",
			PID:  "rrr",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(lstore *mockdb.MockLoadBalancerInterface, pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Loadbalancer_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(lstore *mockdb.MockLoadBalancerInterface, pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
				lstore.EXPECT().Find(gomock.Any(), uint64(loadbalancer.ID)).Times(1).Return(&models.LoadBalancer{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			pstore := mockdb.NewMockIProject(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			lstore := mockdb.NewMockLoadBalancerInterface(ctrl)
			ostore := mockdb.NewMockOrganizationInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(lstore, pstore, ustore)
			tLoadbalancerServer := &LoadbalancerTestServer{
				MUser:         ustore,
				MProject:      pstore,
				MLoadBalancer: lstore,
				MOrg:          ostore,
			}
			pserver := newLoadbalancerTestServer(t, tLoadbalancerServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/loadbalancer/%v", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			pserver.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}

func TestGetLoadBalancerStatusAPI(t *testing.T) {
	tProject := &models.Project{
		Name:           "test-pro",
		Description:    "testing purpose",
		ProjectCode:    "TTTTT",
		UserID:         1,
		SubscriptionID: 1,
	}
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		IsAdmin:   true,
		Company:   "test",
	}
	user.ID = 1
	tProject.ID = 1
	loadbalancer := &models.LoadBalancer{
		Name:         "nnn",
		CustomDomain: "uuu",
		ProjectID:    1,
	}
	loadbalancer.ID = 1
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(lstore *mockdb.MockLoadBalancerInterface, pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		// {
		// 	name: "OK",
		// 	PID:  "1",
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
		// 	},
		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Ema).Times(1).Return(user, nil)
		// 	},
		// 	buildStubs: func(lstore *mockdb.MockLoadBalancerInterface, pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
		// 		lstore.EXPECT().
		// 			Find(gomock.Any(), gomock.Eq(uint64(loadbalancer.ID))).
		// 			Times(1).
		// 			Return(loadbalancer, nil)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusOK, recorder.Code)
		// 	},
		// },
		{
			name: "OKError",
			PID:  "rrr",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(lstore *mockdb.MockLoadBalancerInterface, pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Loadbalancer_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(lstore *mockdb.MockLoadBalancerInterface, pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
				lstore.EXPECT().Find(gomock.Any(), uint64(loadbalancer.ID)).Times(1).Return(&models.LoadBalancer{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			pstore := mockdb.NewMockIProject(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			lstore := mockdb.NewMockLoadBalancerInterface(ctrl)
			ostore := mockdb.NewMockOrganizationInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(lstore, pstore, ustore)
			tLoadbalancerServer := &LoadbalancerTestServer{
				MUser:         ustore,
				MProject:      pstore,
				MLoadBalancer: lstore,
				MOrg:          ostore,
			}
			pserver := newLoadbalancerTestServer(t, tLoadbalancerServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/loadbalancer/%v/status", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			pserver.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}
