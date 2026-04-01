package controllers

import (
	mockdb "01cloud-api/api/controllers/mocks"
	"01cloud-api/api/middlewares"
	"01cloud-api/api/models"
	"01cloud-api/api/utils/cache"
	"bytes"
	"database/sql"
	"encoding/json"
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

type OrginizationPlanServer struct {
	MUser       models.UserInterface
	MockOrgPlan models.OrganizationPlanInterface
}

func newOrgPlanTestServer(t *testing.T, store *OrginizationPlanServer) *Server {
	server, err := NewOrgPlanServer(store)
	require.NoError(t, err)
	return server
}

func NewOrgPlanServer(store *OrginizationPlanServer) (*Server, error) {
	server := &Server{
		MockInterface: store,
		Router:        mux.NewRouter(),
		Cache:         cache.NewCache(1 * time.Minute),
	}
	if oserver, ok := server.MockInterface.(*OrginizationPlanServer); ok {
		orgPlanInterface = oserver.MockOrgPlan
		middlewares.UserI = oserver.MUser
	}
	server.initializeRoutes("http://api.example.com")
	return server, nil
}

func TestCreateOrganizationPlanAPI(t *testing.T) {
	data := &models.OrganizationPlan{
		Name:     "test",
		Cluster:  50,
		Memory:   5000,
		Cores:    1000,
		NoOfUser: 5,
		Price:    10,
		Weight:   10,
		Active:   true,
	}
	data.ID = 1
	user := &models.User{
		FirstName:     "test1",
		LastName:      "test2",
		Image:         "test.jpg",
		Company:       "testcomp",
		Designation:   "test",
		Password:      "test123",
		Email:         "bish@gmail.com",
		EmailVerified: true,
		Active:        true,
		IsAdmin:       true,
	}
	user.ID = 1
	testCases := []struct {
		name          string
		body          map[string]interface{}
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(ostore *mockdb.MockOrganizationPlanInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"name":       "test",
				"cluster":    50,
				"cores":      1000,
				"memory":     5999,
				"price":      10,
				"no_of_user": 2,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(data, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name: "UNAUTHORIZED",
			body: map[string]interface{}{},
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "IO-UTIL-ERROR",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "UNPROCESSABLE-ENITITY",
			body: map[string]interface{}{
				"name": 12345,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "VALIDATE_ERROR",
			body: map[string]interface{}{
				"name": "",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "SAVE_INTERNAL_SERVER_ERROR",
			body: map[string]interface{}{
				"name":       "test",
				"cluster":    50,
				"cores":      1000,
				"memory":     5999,
				"price":      10,
				"no_of_user": 2,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(&models.OrganizationPlan{}, sql.ErrNoRows)
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
			ostore := mockdb.NewMockOrganizationPlanInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(ostore)    //passed mock stubs
			orgPlanServer := &OrginizationPlanServer{
				MUser:       ustore,
				MockOrgPlan: ostore,
			}
			oserver := newOrgPlanTestServer(t, orgPlanServer) //initilize test server
			recorder := httptest.NewRecorder()
			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)
			var request *http.Request
			url := "/organizationPlan"
			if tc.name == "IO-UTIL-ERROR" {
				request, err = http.NewRequest(http.MethodPost, url, errReader(0))
			} else {
				request, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			}
			assert.NoError(t, err)
			tc.setupAuth(t, request)
			oserver.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestOrgPlanAPI(t *testing.T) {
	data := &models.OrganizationPlan{
		Name:     "test",
		Cluster:  50,
		Memory:   5000,
		Cores:    1000,
		NoOfUser: 5,
		Price:    10,
		Weight:   10,
		Active:   true,
	}
	data.ID = 1
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
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(ostore *mockdb.MockOrganizationPlanInterface, ustore *mockdb.MockUserInterface)
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
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface, ustore *mockdb.MockUserInterface) {
				ostore.EXPECT().Find(gomock.Any(), gomock.Eq(uint64(data.ID))).Times(1).Return(data, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
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
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface, ustore *mockdb.MockUserInterface) {
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
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface, ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ORG_PLAN_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface, ustore *mockdb.MockUserInterface) {
				ostore.EXPECT().Find(gomock.Any(), gomock.Eq(uint64(data.ID))).Times(1).Return(&models.OrganizationPlan{}, sql.ErrNoRows)
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
			pstore := mockdb.NewMockOrganizationPlanInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(pstore, ustore)
			orgPlanServer := &OrginizationPlanServer{
				MUser:       ustore,
				MockOrgPlan: pstore,
			}
			oserver := newOrgPlanTestServer(t, orgPlanServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/organizationPlan/%v", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			oserver.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateteOrgPlanAPI(t *testing.T) {
	user := &models.User{
		FirstName:     "test1",
		LastName:      "test2",
		Image:         "test.jpg",
		Company:       "testcomp",
		Designation:   "test",
		Password:      "test123",
		Email:         "bish@gmail.com",
		EmailVerified: true,
		Active:        true,
		IsAdmin:       true,
	}
	user.ID = 1
	data := &models.OrganizationPlan{
		Name:     "test",
		Cluster:  50,
		Memory:   5000,
		Cores:    1000,
		NoOfUser: 5,
		Price:    10,
		Weight:   10,
		Active:   true,
	}
	data.ID = 1
	testCases := []struct {
		name          string
		PID           string
		body          map[string]interface{}
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(ostore *mockdb.MockOrganizationPlanInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			body: map[string]interface{}{
				"name":    "test",
				"cluster": 50,
				"cores":   1000,
				"memory":  5999,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().Find(gomock.Any(), uint64(data.ID)).Times(1).Return(data, nil)
				ostore.EXPECT().Update(gomock.Any(), gomock.Any()).Times(1).Return(data, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UNAUTHORIZED",
			PID:  "1",
			body: map[string]interface{}{},
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "BAD_REQUEST",
			PID:  "A",
			body: map[string]interface{}{},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ORG_PLAN_NOT_FOUND",
			PID:  "1",
			body: map[string]interface{}{},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().Find(gomock.Any(), uint64(data.ID)).Times(1).Return(&models.OrganizationPlan{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "IO-UTIL-ERROR",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().Find(gomock.Any(), uint64(data.ID)).Times(1).Return(data, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "INVALID_FORMAT",
			PID:  "1",
			body: map[string]interface{}{
				"name": true,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().Find(gomock.Any(), uint64(data.ID)).Times(1).Return(data, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "UPDATE_INTERNAL_SERVER_ERROR",
			PID:  "1",
			body: map[string]interface{}{
				"name": "test_group",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().Find(gomock.Any(), uint64(data.ID)).Times(1).Return(data, nil)
				ostore.EXPECT().Update(gomock.Any(), gomock.Any()).Times(1).Return(&models.OrganizationPlan{}, sql.ErrNoRows)
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
			ostore := mockdb.NewMockOrganizationPlanInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(ostore)    //passed mock stubs
			orgPlanServer := &OrginizationPlanServer{
				MUser:       ustore,
				MockOrgPlan: ostore,
			}
			oserver := newOrgPlanTestServer(t, orgPlanServer) //initilize test server
			recorder := httptest.NewRecorder()
			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)
			var request *http.Request
			url := fmt.Sprintf("/organizationPlan/%v", tc.PID)
			if tc.name == "IO-UTIL-ERROR" {
				request, err = http.NewRequest(http.MethodPut, url, errReader(0))
			} else {
				request, err = http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			}
			assert.NoError(t, err)
			tc.setupAuth(t, request)
			oserver.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDELETEORGPLANAPI(t *testing.T) {
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		IsAdmin:   true,
		Company:   "test",
	}
	user.ID = 1
	data := &models.OrganizationPlan{
		Name:     "test",
		Cluster:  50,
		Memory:   5000,
		Cores:    1000,
		NoOfUser: 5,
		Price:    10,
		Weight:   10,
		Active:   true,
	}
	data.ID = 1

	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(ostore *mockdb.MockOrganizationPlanInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().Find(gomock.Any(), uint64(data.ID)).Times(1).Return(data, nil)
				ostore.EXPECT().Delete(gomock.Any(), uint64(data.ID)).Times(1).Return(int64(1), nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UNAUTHORIZED",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "BAD-REQUEST",
			PID:  "A",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ORGP/organizationPlans_LAN_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().Find(gomock.Any(), uint64(data.ID)).Times(1).Return(&models.OrganizationPlan{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "DELETE_ERROR",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), 0, user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().Find(gomock.Any(), uint64(data.ID)).Times(1).Return(data, nil)
				ostore.EXPECT().Delete(gomock.Any(), uint64(data.ID)).Times(1).Return(int64(0), sql.ErrNoRows)
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
			ostore := mockdb.NewMockOrganizationPlanInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(ostore)    //passed mock stubs
			orgPlanServer := &OrginizationPlanServer{
				MUser:       ustore,
				MockOrgPlan: ostore,
			}
			oserver := newOrgPlanTestServer(t, orgPlanServer) //initilize test server
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/organizationPlan/%v", tc.PID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, request) //passed authorization token

			oserver.Router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder) //check response
		})
	}
}

func TestGetAllOrgPlanAPI(t *testing.T) {
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		IsAdmin:   true,
		Company:   "test",
	}
	user.ID = 1
	datas := &[]models.OrganizationPlan{
		{
			Name:     "test",
			Cluster:  50,
			Memory:   5000,
			Cores:    1000,
			NoOfUser: 5,
			Price:    10,
			Weight:   10,
			Active:   true,
		},
		{
			Name:     "test1",
			Cluster:  60,
			Memory:   8000,
			Cores:    2000,
			NoOfUser: 10,
			Price:    20,
			Weight:   20,
			Active:   true,
		},
	}
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(ostore *mockdb.MockOrganizationPlanInterface)
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
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().FindAll(gomock.Any()).Times(1).Return(datas, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UNAUTHORIZED",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ORGANIZATION_PLAN_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().FindAll(gomock.Any()).Times(1).Return(&[]models.OrganizationPlan{}, sql.ErrNoRows)
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
			ostore := mockdb.NewMockOrganizationPlanInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(ostore)    //passed mock stubs
			orgPlanServer := &OrginizationPlanServer{
				MUser:       ustore,
				MockOrgPlan: ostore,
			}
			oserver := newOrgPlanTestServer(t, orgPlanServer)
			recorder := httptest.NewRecorder()

			url := "/organizationPlans"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			oserver.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}

func TestGetADMINAllOrgPlanAPI(t *testing.T) {
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		IsAdmin:   true,
		Company:   "test",
	}
	user.ID = 1
	datas := &[]models.OrganizationPlan{
		{
			Name:     "test",
			Cluster:  50,
			Memory:   5000,
			Cores:    1000,
			NoOfUser: 5,
			Price:    10,
			Weight:   10,
			Active:   true,
		},
		{
			Name:     "test1",
			Cluster:  60,
			Memory:   8000,
			Cores:    2000,
			NoOfUser: 10,
			Price:    20,
			Weight:   20,
			Active:   true,
		},
	}
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(ostore *mockdb.MockOrganizationPlanInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().FindAllWithInActive(gomock.Any()).Times(1).Return(datas, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UNAUTHORIZED",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ORGANIZATION_PLAN_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ostore *mockdb.MockOrganizationPlanInterface) {
				ostore.EXPECT().FindAllWithInActive(gomock.Any()).Times(1).Return(&[]models.OrganizationPlan{}, sql.ErrNoRows)
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
			ostore := mockdb.NewMockOrganizationPlanInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(ostore)    //passed mock stubs
			orgPlanServer := &OrginizationPlanServer{
				MUser:       ustore,
				MockOrgPlan: ostore,
			}
			oserver := newOrgPlanTestServer(t, orgPlanServer)
			recorder := httptest.NewRecorder()

			url := "/admin/organizationPlans"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			oserver.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}
