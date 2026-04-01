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

type UserRoleTServer struct {
	MUser     models.UserInterface
	MUserRole models.UserRoleInterface
	MOrg      models.OrganizationInterface
	Router    *mux.Router
}

func NewUserRoleTServer(t *testing.T, store *UserRoleTServer) *Server {
	server, err := newUserRoleServer(store)
	require.NoError(t, err)
	return server
}

func newUserRoleServer(store *UserRoleTServer) (*Server, error) {
	server := &Server{
		MockInterface: store,
		Router:        mux.NewRouter(),
		Cache:         cache.NewCache(1 * time.Minute),
	}
	if gserver, ok := server.MockInterface.(*UserRoleTServer); ok {
		userroleInterface = gserver.MUserRole
		userInterface = gserver.MUser
		orgInterface = gserver.MOrg
		middlewares.UserI = gserver.MUser
	}
	server.initializeRoutes("http://api.example.com")
	return server, nil
}

func TestCreateUserRoleAPI(t *testing.T) {
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
	org := &models.Organization{
		Name: "testorg",
	}
	org.ID = 1

	UserRole := &models.UserRole{
		Name:        "test",
		Code:        112,
		Description: "good",
	}
	UserRole.ID = 1
	// UserRoles := &[]models.UserRole{
	// 	{
	// 		Name: "test",
	// 	},
	// }
	testCases := []struct {
		name          string
		body          map[string]interface{}
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(dstore *mockdb.MockUserRoleInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"name":        "test_UserRole",
				"code":        112,
				"description": "good",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				//gstore.EXPECT().FindAll(gomock.Any(), uint64(org.ID), gomock.Any()).Times(1).Return(UserRoles, nil)
				gstore.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(UserRole, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		// {
		// 	name: "UNAUTHORIZED",
		// 	body: map[string]interface{}{},
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 	},
		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
		// 	},
		// 	buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		// 	},
		// },
		{
			name: "IO-UTIL-ERROR",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
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
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "VALIDATION_DATA",
			body: map[string]interface{}{
				"name": "aa",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "SAVE_INTERNAL_SERVER_ERROR",
			body: map[string]interface{}{
				"name":        "test_UserRole",
				"code":        1122,
				"description": "aaa",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				gstore.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(&models.UserRole{}, sql.ErrNoRows)
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
			gstore := mockdb.NewMockUserRoleInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(gstore)    //passed mock stubs
			tUserRoleserver := &UserRoleTServer{
				MUserRole: gstore,
				MUser:     ustore,
			}
			server := NewUserRoleTServer(t, tUserRoleserver) //initilize test server
			recorder := httptest.NewRecorder()
			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)
			var request *http.Request
			url := "/role"
			if tc.name == "IO-UTIL-ERROR" {
				request, err = http.NewRequest(http.MethodPost, url, errReader(0))
			} else {
				request, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			}
			assert.NoError(t, err)
			tc.setupAuth(t, request)
			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGETUserRoleAPI(t *testing.T) {
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		IsAdmin:   true,
		Company:   "test",
	}
	user.ID = 1
	org := &models.Organization{
		Name: "testorg",
	}
	org.ID = 1

	UserRole := &models.UserRole{
		Name:        "domainhem",
		Code:        112,
		Description: "good",
	}

	UserRole.ID = 1
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(dstore *mockdb.MockUserRoleInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserRole.ID)).Times(1).Return(UserRole, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "BAD-REQUEST",
			PID:  "A",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(dstore *mockdb.MockUserRoleInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserRole.ID)).Times(1).Return(&models.UserRole{}, sql.ErrNoRows)

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
			gstore := mockdb.NewMockUserRoleInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(gstore)    //passed mock stubs
			tUserRoleserver := &UserRoleTServer{
				MUserRole: gstore,
				MUser:     ustore,
			}
			server := NewUserRoleTServer(t, tUserRoleserver)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/role/%v", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, request) //passed authorization token

			server.Router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder) //check response
		})
	}
}

func TestGETUserRoleLISTAPI(t *testing.T) {
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		IsAdmin:   true,
		Company:   "test",
	}
	user.ID = 1
	org := &models.Organization{
		Name: "testorg",
	}
	org.ID = 1

	UserRole := &[]models.UserRole{
		{
			Name:        "test",
			Code:        112,
			Description: "good",
		},
	}
	//possible cases
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(dstore *mockdb.MockUserRoleInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				gstore.EXPECT().FindAll(gomock.Any()).Times(1).Return(UserRole, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				gstore.EXPECT().FindAll(gomock.Any()).Times(1).Return(&[]models.UserRole{}, sql.ErrNoRows)
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
			gstore := mockdb.NewMockUserRoleInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(gstore)    //passed mock stubs
			tUserRoleserver := &UserRoleTServer{
				MUserRole: gstore,
				MUser:     ustore,
			}
			server := NewUserRoleTServer(t, tUserRoleserver) //initilize test server
			recorder := httptest.NewRecorder()

			url := "/roles"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			//request.FormValue(tc.query)
			assert.NoError(t, err)

			tc.setupAuth(t, request) //passed authorization token

			server.Router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder) //check response
		})
	}
}

func TestUpdateteUserRoleAPI(t *testing.T) {
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
	org := &models.Organization{
		Name: "testorg",
	}
	org.ID = 1

	UserRole := &models.UserRole{
		Name:        "test",
		Code:        112,
		Description: "good",
	}
	UserRole.ID = 1
	testCases := []struct {
		name          string
		PID           string
		body          map[string]interface{}
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(dstore *mockdb.MockUserRoleInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			body: map[string]interface{}{
				"name":        "test",
				"code":        112,
				"description": "good",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserRole.ID)).Times(1).Return(UserRole, nil)
				gstore.EXPECT().Update(gomock.Any(), gomock.Any()).Times(1).Return(UserRole, nil)
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
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UNPROCESSABLE-ENITITY",
			PID:  "A",
			body: map[string]interface{}{},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "IO-UTIL-ERROR",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserRole.ID)).Times(1).Return(UserRole, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "UserRole_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserRole.ID)).Times(1).Return(&models.UserRole{}, sql.ErrNoRows)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "INVALID_FORMAT",
			PID:  "1",
			body: map[string]interface{}{
				"name": true,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserRole.ID)).Times(1).Return(UserRole, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "UPDATE_INTERNAL_SERVER_ERROR",
			PID:  "1",
			body: map[string]interface{}{
				"name":        "test",
				"code":        112,
				"description": "good",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserRole.ID)).Times(1).Return(UserRole, nil)
				gstore.EXPECT().Update(gomock.Any(), gomock.Any()).Times(1).Return(&models.UserRole{}, sql.ErrNoRows)
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
			gstore := mockdb.NewMockUserRoleInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(gstore)    //passed mock stubs
			tUserRoleserver := &UserRoleTServer{
				MUserRole: gstore,
				MUser:     ustore,
			}
			server := NewUserRoleTServer(t, tUserRoleserver) //initilize test server
			recorder := httptest.NewRecorder()
			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)
			var request *http.Request
			url := fmt.Sprintf("/role/%v", tc.PID)
			if tc.name == "IO-UTIL-ERROR" {
				request, err = http.NewRequest(http.MethodPut, url, errReader(0))
			} else {
				request, err = http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			}
			assert.NoError(t, err)
			tc.setupAuth(t, request)
			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDELETEUserRoleAPI(t *testing.T) {
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		IsAdmin:   true,
		Company:   "test",
	}
	user.ID = 1
	org := &models.Organization{
		Name: "testorg",
	}
	org.ID = 1

	UserRole := &models.UserRole{
		Name:        "domainhem",
		Code:        112,
		Description: "good",
	}

	UserRole.ID = 1
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(dstore *mockdb.MockUserRoleInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserRole.ID)).Times(1).Return(UserRole, nil)
				gstore.EXPECT().Delete(gomock.Any(), uint64(UserRole.ID)).Times(1).Return(int64(UserRole.ID), nil)
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
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "BAD-REQUEST",
			PID:  "A",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(dstore *mockdb.MockUserRoleInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UserRole_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserRole.ID)).Times(1).Return(&models.UserRole{}, sql.ErrNoRows)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "DELETE_ERROR",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockUserRoleInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserRole.ID)).Times(1).Return(UserRole, nil)
				gstore.EXPECT().Delete(gomock.Any(), uint64(UserRole.ID)).Times(1).Return(int64(0), sql.ErrNoRows)
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
			gstore := mockdb.NewMockUserRoleInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(gstore)    //passed mock stubs
			tUserRoleserver := &UserRoleTServer{
				MUserRole: gstore,
				MUser:     ustore,
			}
			server := NewUserRoleTServer(t, tUserRoleserver)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/role/%v", tc.PID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, request) //passed authorization token

			server.Router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder) //check response
		})
	}
}
