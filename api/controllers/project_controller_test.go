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

type ProjectTestServer struct {
	MUser         models.UserInterface
	MProject      models.IProject
	MSubscription models.SubscriptionInterface
	MOrg          models.OrganizationInterface
	MActivity     models.ActivityInterface
}

func newProjectTestServer(t *testing.T, store *ProjectTestServer) *Server {
	server, err := NewProjectServer(store)
	require.NoError(t, err)
	return server
}

func NewProjectServer(store *ProjectTestServer) (*Server, error) {
	server := &Server{
		MockInterface: store,
		Router:        mux.NewRouter(),
		Cache:         cache.NewCache(1 * time.Minute),
	}
	if pserver, ok := server.MockInterface.(*ProjectTestServer); ok {
		iproject = pserver.MProject
		userInterface = pserver.MUser
		subscriptionInterface = pserver.MSubscription
		orgInterface = pserver.MOrg
		activityInterface = pserver.MActivity
		middlewares.UserI = pserver.MUser
	}
	server.initializeRoutes("http://api.example.com")
	return server, nil
}

// func TestCreateProjectAPI(t *testing.T) {
// 	// activity := &models.Activity{
// 	// 	Action:    "aa",
// 	// 	Module:    "xxz",
// 	// 	Active:    true,
// 	// 	Remarks:   "Nepal",
// 	// 	ProjectID: 1,
// 	// }
// 	subscription := &models.Subscription{
// 		Name:         "Basic Plan",
// 		Apps:         5,
// 		DiskSpace:    5,
// 		Memory:       2,
// 		Cores:        1,
// 		DataTransfer: 5,
// 		Price:        100,
// 		Weight:       1,
// 		Backups:      5,
// 		Active:       true,
// 	}
// 	subscription.ID = 1
// 	tProject := &models.Project{
// 		Name:           "test",
// 		UserID:         1,
// 		SubscriptionID: 1,
// 		OrganizationId: 1,
// 		Active:         true,
// 	}
// 	tProject.ID = 1
// 	user := &models.User{
// 		FirstName:     "test1",
// 		LastName:      "test2",
// 		Image:         "test.jpg",
// 		Company:       "testcomp",
// 		Designation:   "test",
// 		Password:      "test123",
// 		Email:         "bish@gmail.com",
// 		EmailVerified: true,
// 		Active:        true,
// 		IsAdmin:       true,
// 	}
// 	user.ID = 1
// 	org := &models.Organization{
// 		Name:   "testorg",
// 		UserID: 1,
// 	}
// 	org.ID = 1
// 	testCases := []struct {
// 		name          string
// 		body          map[string]interface{}
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(dstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface, sstore *mockdb.MockSubscriptionInterface, ostore *mockdb.MockOrganizationInterface, astore *mockdb.MockActivityInterface)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		// {
// 		// 	name: "OK",
// 		// 	body: map[string]interface{}{
// 		// 		"name":            "test-project",
// 		// 		"subscription_id": 1,
// 		// 		"active":          true,
// 		// 		"organization_id": 1,
// 		// 	},
// 		// 	setupAuth: func(t *testing.T, request *http.Request) {
// 		// 		addAuthorization(t, request, uint(user.ID), uint(tProject.OrganizationId))
// 		// 	},
// 		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 		// 	},
// 		// 	buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface, sstore *mockdb.MockSubscriptionInterface, ostore *mockdb.MockOrganizationInterface, astore *mockdb.MockActivityInterface) {
// 		// 		sstore.EXPECT().Find(gomock.Any(), uint64(1)).Times(1).Return(subscription, nil)
// 		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(), gomock.Eq(user.Email)).Times(1).Return(user, nil)
// 		// 		pstore.EXPECT().IsNameExists(gomock.Any(), uint(user.ID), "test-project", gomock.Any()).Times(1).Return(false)
// 		// 		ostore.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(org, nil)
// 		// 		pstore.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(tProject, nil)
// 		// 		//astore.EXPECT().SaveActivityWithJson(gomock.Any(),gomock.Any()).Times(1).Return()
// 		// 		astore.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(activity, nil)

// 		// 	},
// 		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
// 		// 		assert.Equal(t, http.StatusOK, recorder.Code)
// 		// 	},
// 		// },
// 		{
// 			name: "OK",
// 			body: map[string]interface{}{
// 				"name":            "test-project",
// 				"subscription_id": 1,
// 				"active":          true,
// 				"organization_id": 1,
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(tProject.OrganizationId))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface, sstore *mockdb.MockSubscriptionInterface, ostore *mockdb.MockOrganizationInterface, astore *mockdb.MockActivityInterface) {
// 				sstore.EXPECT().Find(gomock.Any(), uint64(1)).Times(1).Return(subscription, nil)
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), gomock.Eq(user.Email)).Times(1).Return(user, nil)
// 				pstore.EXPECT().IsNameExists(gomock.Any(), uint(user.ID), "test-project", gomock.Any()).Times(1).Return(false)
// 				ostore.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(org, nil)
// 				pstore.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(&models.Project{}, sql.ErrNoRows)
// 				//astore.EXPECT().SaveActivityWithJson(gomock.Any(),gomock.Any()).Times(1).Return()
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "IO-UTIL-ERROR",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(tProject.OrganizationId))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface, sstore *mockdb.MockSubscriptionInterface, ostore *mockdb.MockOrganizationInterface, astore *mockdb.MockActivityInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNPROCESSABLE-ENITITY",
// 			body: map[string]interface{}{
// 				"name":            "test",
// 				"active":          "true",
// 				"organization_id": "2",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(tProject.OrganizationId))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(dstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface, sstore *mockdb.MockSubscriptionInterface, ostore *mockdb.MockOrganizationInterface, astore *mockdb.MockActivityInterface) {

// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "VALIDATE_ERROR",
// 			body: map[string]interface{}{
// 				"name":            "",
// 				"active":          true,
// 				"organization_id": 1,
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(tProject.OrganizationId))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(dstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface, sstore *mockdb.MockSubscriptionInterface, ostore *mockdb.MockOrganizationInterface, astore *mockdb.MockActivityInterface) {

// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "SUBSCRIPTION_NOT_FOUND",
// 			body: map[string]interface{}{
// 				"name":            "test-project",
// 				"subscription_id": 1,
// 				"active":          true,
// 				"organization_id": 1,
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(tProject.OrganizationId))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(dstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface, sstore *mockdb.MockSubscriptionInterface, ostore *mockdb.MockOrganizationInterface, astore *mockdb.MockActivityInterface) {
// 				sstore.EXPECT().Find(gomock.Any(), uint64(1)).Times(1).Return(&models.Subscription{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNAUTHORIZED",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(dstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface, sstore *mockdb.MockSubscriptionInterface, ostore *mockdb.MockOrganizationInterface, astore *mockdb.MockActivityInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "USER_NOT_FOUND",
// 			body: map[string]interface{}{
// 				"name":            "test-project",
// 				"subscription_id": 1,
// 				"active":          true,
// 				"organization_id": 1,
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(tProject.OrganizationId))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(dstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface, sstore *mockdb.MockSubscriptionInterface, ostore *mockdb.MockOrganizationInterface, astore *mockdb.MockActivityInterface) {
// 				sstore.EXPECT().Find(gomock.Any(), uint64(1)).Times(1).Return(subscription, nil)
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(&models.User{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "USER_NOT_FOUND",
// 			body: map[string]interface{}{
// 				"name":            "test-project",
// 				"subscription_id": 1,
// 				"active":          true,
// 				"organization_id": 1,
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(tProject.OrganizationId))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(dstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface, sstore *mockdb.MockSubscriptionInterface, ostore *mockdb.MockOrganizationInterface, astore *mockdb.MockActivityInterface) {
// 				sstore.EXPECT().Find(gomock.Any(), uint64(1)).Times(1).Return(subscription, nil)
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), gomock.Eq(user.Email)).Times(1).Return(user, nil)
// 				dstore.EXPECT().IsNameExists(gomock.Any(), uint(user.ID), "test-project", gomock.Any()).Times(1).Return(true)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "ORG_NOT_FOUND",
// 			body: map[string]interface{}{
// 				"name":            "test-project",
// 				"subscription_id": 1,
// 				"active":          true,
// 				"organization_id": 1,
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(tProject.OrganizationId))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface, sstore *mockdb.MockSubscriptionInterface, ostore *mockdb.MockOrganizationInterface, astore *mockdb.MockActivityInterface) {
// 				sstore.EXPECT().Find(gomock.Any(), uint64(1)).Times(1).Return(subscription, nil)
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), gomock.Eq(user.Email)).Times(1).Return(user, nil)
// 				pstore.EXPECT().IsNameExists(gomock.Any(), uint(user.ID), "test-project", gomock.Any()).Times(1).Return(false)
// 				ostore.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(&models.Organization{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 	}
// 	for i := range testCases {
// 		tc := testCases[i]

// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
// 			pstore := mockdb.NewMockIProject(ctrl)
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			sstore := mockdb.NewMockSubscriptionInterface(ctrl)
// 			ostore := mockdb.NewMockOrganizationInterface(ctrl)
// 			astore := mockdb.NewMockActivityInterface(ctrl)
// 			tc.setMiddleware(ustore) //passed middleware
// 			tc.buildStubs(pstore, ustore, sstore, ostore, astore)
// 			tProjectServer := &ProjectTestServer{
// 				MUser:         ustore,
// 				MProject:      pstore,
// 				MSubscription: sstore,
// 				MOrg:          ostore,
// 				MActivity:     astore,
// 			}
// 			pserver := newProjectTestServer(t, tProjectServer)
// 			recorder := httptest.NewRecorder()
// 			// Marshal body data to JSON
// 			data, err := json.Marshal(tc.body)
// 			assert.NoError(t, err)
// 			var request *http.Request
// 			url := "/project"
// 			if tc.name == "IO-UTIL-ERROR" {
// 				request, err = http.NewRequest(http.MethodPost, url, errReader(0))
// 			} else {
// 				request, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
// 			}
// 			assert.NoError(t, err)
// 			tc.setupAuth(t, request)
// 			pserver.Router.ServeHTTP(recorder, request)
// 			tc.checkResponse(recorder)
// 		})
// 	}
// }

func TestGETProjectAPI(t *testing.T) {
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
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		// {
		// 	name: "OK",
		// 	PID:  "1",
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
		// 	},
		// 	buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
		// 		ustore.EXPECT().
		// 			FindUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
		// 			Times(1).
		// 			Return(user, nil)

		// 		pstore.EXPECT().
		// 			Find(gomock.Any(), gomock.Eq(uint64(tProject.ID))).
		// 			Times(1).
		// 			Return(tProject, nil)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		// 	},
		// },
		{
			name: "BADREQUEST",
			PID:  "A",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UNAUTHORIZED",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
			},
			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "USER_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), gomock.Eq(user.ID)).Times(1).Return(&models.User{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "PROJECT_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), gomock.Eq(user.ID)).Times(1).Return(user, nil)
				pstore.EXPECT().Find(gomock.Any(), gomock.Eq(uint64(tProject.ID))).Times(1).Return(&models.Project{}, sql.ErrNoRows)
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
			tc.setMiddleware(ustore)
			tc.buildStubs(pstore, ustore)
			tProjectServer := &ProjectTestServer{
				MUser:    ustore,
				MProject: pstore,
			}
			pserver := newProjectTestServer(t, tProjectServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/project/%v", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			pserver.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}

func TestGetProjectByOrganizationForAdminAPI(t *testing.T) {
	orgID := 1
	tProject := &[]models.Project{
		{
			Name:           "test-pro",
			Description:    "testing purpose",
			ProjectCode:    "TTTTT",
			UserID:         1,
			SubscriptionID: 1,
			OrganizationId: 1,
		},
		{
			Name:           "test-pro1",
			Description:    "testing purpose",
			ProjectCode:    "TTTTT",
			UserID:         1,
			SubscriptionID: 1,
			OrganizationId: 1,
		},
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
	testCases := []struct {
		name          string
		OID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(pstore *mockdb.MockIProject)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			OID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(orgID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(pstore *mockdb.MockIProject) {
				pstore.EXPECT().FindProject(gomock.Any(), uint(orgID)).Times(1).Return(tProject, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		// {
		// 	name: "BADREQUEST",
		// 	OID:  "A",
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(orgID))
		// 	},
		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
		// 		ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
		// 	},
		// 	buildStubs: func(pstore *mockdb.MockIProject) {
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "PROJECT_NOT_FOUND",
		// 	OID:  "1",
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(orgID))
		// 	},
		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
		// 		ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
		// 	},
		// 	buildStubs: func(pstore *mockdb.MockIProject) {
		// 		pstore.EXPECT().FindProject(gomock.Any(), uint(orgID)).Times(1).Return(&[]models.Project{}, sql.ErrNoRows)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusNotFound, recorder.Code)
		// 	},
		// },
	}
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			pstore := mockdb.NewMockIProject(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(pstore)
			tProjectServer := &ProjectTestServer{
				MUser:    ustore,
				MProject: pstore,
			}
			pserver := newProjectTestServer(t, tProjectServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/admin/project/%v", tc.OID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			pserver.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}

// func TestGetProjectOfUserOnlyAPI(t *testing.T) {
// 	orgID := 1
// 	user := &models.User{
// 		FirstName: "test1",
// 		LastName:  "test1",
// 		Email:     "test@gmai.com",
// 		Active:    true,
// 		IsAdmin:   true,
// 		Company:   "test",
// 	}
// 	user.ID = 1
// 	testCases := []struct {
// 		name          string
// 		OID           string
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(pstore *mockdb.MockIProject)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "OK",
// 			OID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(orgID))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(pstore *mockdb.MockIProject) {
// 				pstore.EXPECT().FindUserProjectOnly(gomock.Any(), uint(orgID), gomock.Any()).Times(1).Return([]map[string]interface{}{
// 					{"test": "app"},
// 				}, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNAUTHORIZED",
// 			OID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {

// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(pstore *mockdb.MockIProject) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "BADREQUEST",
// 			OID:  "A",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(orgID))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(pstore *mockdb.MockIProject) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "PROJECT_NOT_FOUND",
// 			OID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(orgID))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(pstore *mockdb.MockIProject) {
// 				pstore.EXPECT().FindUserProjectOnly(gomock.Any(), uint(orgID), gomock.Any()).Times(1).Return(nil, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
// 			},
// 		},
// 	}
// 	for i := range testCases {
// 		tc := testCases[i]

// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
// 			pstore := mockdb.NewMockIProject(ctrl)
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			tc.setMiddleware(ustore)
// 			tc.buildStubs(pstore)
// 			tProjectServer := &ProjectTestServer{
// 				MUser:    ustore,
// 				MProject: pstore,
// 			}
// 			pserver := newProjectTestServer(t, tProjectServer)
// 			recorder := httptest.NewRecorder()

// 			url := fmt.Sprintf("/user/%v/projects", tc.OID)
// 			request, err := http.NewRequest(http.MethodGet, url, nil)
// 			assert.NoError(t, err)
// 			tc.setupAuth(t, request)

// 			pserver.Router.ServeHTTP(recorder, request)
// 			//check response
// 			tc.checkResponse(recorder)
// 		})
// 	}
// }

func TestGetProjectsAPI(t *testing.T) {
	orgID := 1
	tProject := []models.Project{
		{
			Name:           "test-pro",
			Description:    "testing purpose",
			ProjectCode:    "TTTTT",
			UserID:         1,
			SubscriptionID: 1,
			OrganizationId: 1,
		},
		{
			Name:           "test-pro1",
			Description:    "testing purpose",
			ProjectCode:    "TTTTT",
			UserID:         1,
			SubscriptionID: 1,
			OrganizationId: 1,
		},
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
	testCases := []struct {
		name          string
		OID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(pstore *mockdb.MockIProject)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			OID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(orgID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(pstore *mockdb.MockIProject) {
				pstore.EXPECT().FindAllByUser(gomock.Any(), uint(user.ID), gomock.Any()).Times(1).Return(tProject, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UNAUTHORIZED",
			OID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {

			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
			},
			buildStubs: func(pstore *mockdb.MockIProject) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "INTERNAL_SERVER_ERROR",
			OID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(orgID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(pstore *mockdb.MockIProject) {
				pstore.EXPECT().FindAllByUser(gomock.Any(), uint(user.ID), gomock.Any()).Times(1).Return([]models.Project{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
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
			tc.setMiddleware(ustore)
			tc.buildStubs(pstore)
			tProjectServer := &ProjectTestServer{
				MUser:    ustore,
				MProject: pstore,
			}
			pserver := newProjectTestServer(t, tProjectServer)
			recorder := httptest.NewRecorder()

			url := "/projects"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			pserver.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}

func TestGetGlobalVariablesAPI(t *testing.T) {
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
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		// {
		// 	name: "OK",
		// 	PID:  "1",
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
		// 	},
		// 	buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
		// 		ustore.EXPECT().
		// 			FindUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
		// 			Times(1).
		// 			Return(user, nil)

		// 		pstore.EXPECT().
		// 			Find(gomock.Any(), gomock.Eq(uint64(tProject.ID))).
		// 			Times(1).
		// 			Return(tProject, nil)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		// 	},
		// },
		{
			name: "BADREQUEST",
			PID:  "A",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UNAUTHORIZED",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
			},
			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "USER_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().
					FindUserByID(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(&models.User{}, sql.ErrNoRows)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "PROJECT_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().
					FindUserByID(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				pstore.EXPECT().
					Find(gomock.Any(), gomock.Eq(uint64(tProject.ID))).
					Times(1).
					Return(&models.Project{}, sql.ErrNoRows)
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
			tc.setMiddleware(ustore)
			tc.buildStubs(pstore, ustore)
			tProjectServer := &ProjectTestServer{
				MUser:    ustore,
				MProject: pstore,
			}
			pserver := newProjectTestServer(t, tProjectServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/project/%v/variables", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			pserver.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}

func TestGetResourceUsedAPI(t *testing.T) {
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
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		// {
		// 	name: "OK",
		// 	PID:  "1",
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
		// 	},
		// 	buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
		// 		ustore.EXPECT().
		// 			FindUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
		// 			Times(1).
		// 			Return(user, nil)

		// 		pstore.EXPECT().
		// 			Find(gomock.Any(), gomock.Eq(uint64(tProject.ID))).
		// 			Times(1).
		// 			Return(tProject, nil)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		// 	},
		// },
		{
			name: "BADREQUEST",
			PID:  "A",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UNAUTHORIZED",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
			},
			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "USER_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().
					FindUserByID(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(&models.User{}, sql.ErrNoRows)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "PROJECT_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(pstore *mockdb.MockIProject, ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().
					FindUserByID(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				pstore.EXPECT().
					Find(gomock.Any(), gomock.Eq(uint64(tProject.ID))).
					Times(1).
					Return(&models.Project{}, sql.ErrNoRows)
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
			tc.setMiddleware(ustore)
			tc.buildStubs(pstore, ustore)
			tProjectServer := &ProjectTestServer{
				MUser:    ustore,
				MProject: pstore,
			}
			pserver := newProjectTestServer(t, tProjectServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/project/%v/resource", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			pserver.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}
