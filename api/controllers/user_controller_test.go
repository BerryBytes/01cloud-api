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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UserTestServer struct {
	MUser models.UserInterface
}

func newUserTestServer(t *testing.T, store *UserTestServer) *Server {
	server, err := NewUserServer(store)
	require.NoError(t, err)
	return server
}

func NewUserServer(store *UserTestServer) (*Server, error) {
	server := &Server{
		MockInterface: store,
		Router:        mux.NewRouter(),
		Cache:         cache.NewCache(1 * time.Minute),
	}
	if userServer, ok := server.MockInterface.(*UserTestServer); ok {
		userInterface = userServer.MUser
		middlewares.UserI = userServer.MUser
	}
	server.initializeRoutes("http://api.example.com")
	return server, nil
}

func TestCreateUserAPI(t *testing.T) {
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
		buildStubs    func(store *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name:       "IO-UTIL-ERROR",
			buildStubs: func(store *mockdb.MockUserInterface) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "UnprocessableEnitity",
			body: map[string]interface{}{
				"first_name": 123,
				"last_name":  "test_1",
				"email":      "test@gmail.com",
				"active":     true,
			},
			buildStubs: func(store *mockdb.MockUserInterface) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "USER_VALIDATE_ERROR",
			body: map[string]interface{}{
				"first_name": "",
				"last_name":  "test_1",
				"email":      "test@gmail.com",
				"password":   "test@123t",
				"active":     true,
			},
			buildStubs: func(store *mockdb.MockUserInterface) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		// {
		// 	name: "USER_SAVE_ERROR",
		// 	body: map[string]interface{}{
		// 		"first_name": "test",
		// 		"last_name":  "test",
		// 		"email":      "test@gmail.com",
		// 		"password":   "test@123t",
		// 		"active":     true,
		// 	},
		// 	buildStubs: func(ustore *mockdb.MockUserInterface) {
		// 		ustore.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Times(1).Return(&models.User{}, sql.ErrNoRows)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
		// 	},
		// },
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.buildStubs(ustore)
			store := &UserTestServer{
				MUser: ustore,
			}
			server := newUserTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)
			var request *http.Request
			url := "/user/register"
			if tc.name == "IO-UTIL-ERROR" {
				request, err = http.NewRequest(http.MethodPost, url, errReader(0))
			} else {
				request, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			}
			assert.NoError(t, err)
			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestChangePasswordAPI(t *testing.T) {
	user := &models.User{
		FirstName:     "test1",
		LastName:      "test2",
		Image:         "test.jpg",
		Company:       "testcomp",
		Designation:   "test",
		Password:      "$2a$10$6MGvHaWSjG4tYghcuxTCQeIiJuyfSFJRvRRSsXF125ld9YHIXQBQ.",
		Email:         "bish@gmail.com",
		EmailVerified: true,
		Active:        true,
	}
	user.ID = 1
	testCases := []struct {
		name          string
		body          map[string]interface{}
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(store *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "IO-UTIL-ERROR",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockUserInterface) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "USER_NOT_FOUND",
			body: map[string]interface{}{
				"new_password":    "123456789",
				"retype_password": "123456789",
				"password":        "1234567890",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(&models.User{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "UNMARSHAL_ERROR",
			body: map[string]interface{}{
				"12345":           123456789,
				"retype_password": "123456789",
				"password":        "1234567890",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "EMPTY_PASSWORD",
			body: map[string]interface{}{
				"new_password":    "",
				"retype_password": "123456789",
				"password":        "1234567890",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "PASSWORD_LENGTH_ERROR",
			body: map[string]interface{}{
				"new_password":    "12345",
				"retype_password": "12345",
				"password":        "1234567890",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "PASSWORD_SAME_AS_OLD_PASSWORD_ERROR",
			body: map[string]interface{}{
				"new_password":    "123456789",
				"retype_password": "123456789",
				"password":        "123456789",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "NEW_PASSWORD_NOT_SAME_AS_RETYPE_PASSWORD_ERROR",
			body: map[string]interface{}{
				"new_password":    "123456789",
				"retype_password": "APPLE123456",
				"password":        "01cl0ud@2o2o",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "NEW_PASSWORD_UPDATE",
			body: map[string]interface{}{
				"new_password":    "123456789",
				"retype_password": "123456789",
				"password":        "$2a$10$6MGvHaWSjG4tYghcuxTCQeIiJuyfSFJRvRRSsXF125ld9YHIXQBQ.",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
				//models.VerifyPassword(user.Password, "01cl0ud@2o2o")
				//user.Password = "123456789"
				//ustore.EXPECT().UpdatePassword(gomock.Any(), user).Times(1).Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(ustore)
			store := &UserTestServer{
				MUser: ustore,
			}
			server := newUserTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)
			var request *http.Request
			url := "/user/change-password"
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

func TestGetUserAPI(t *testing.T) {
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		Company:   "test",
	}
	user.ID = 1

	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(ustore *mockdb.MockUserInterface)
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
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "BAD-REQUEST",
			PID:  "A",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "USER_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(&models.User{}, sql.ErrNoRows)
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
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(ustore)    //passed mock stubs
			testServer := &UserTestServer{
				MUser: ustore,
			}
			server := newUserTestServer(t, testServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/user/%v", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, request) //passed authorization token

			server.Router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder) //check response
		})
	}
}

// func TestGetProfileAPI(t *testing.T) {
// 	user := &models.User{
// 		FirstName: "test1",
// 		LastName:  "test1",
// 		Email:     "test@gmai.com",
// 		Active:    true,
// 		Company:   "test",
// 	}
// 	user.ID = 1

// 	testCases := []struct {
// 		name          string
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(ustore *mockdb.MockUserInterface)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "OK",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "USER_NOT_FOUND",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(&models.User{}, sql.ErrNoRows)
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
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			tc.setMiddleware(ustore) //passed middleware
// 			tc.buildStubs(ustore)    //passed mock stubs
// 			testServer := &UserTestServer{
// 				MUser: ustore,
// 			}
// 			server := newUserTestServer(t, testServer)
// 			recorder := httptest.NewRecorder()

// 			url := "/profile"
// 			request, err := http.NewRequest(http.MethodGet, url, nil)
// 			assert.NoError(t, err)

// 			tc.setupAuth(t, request) //passed authorization token

// 			server.Router.ServeHTTP(recorder, request)

// 			tc.checkResponse(recorder) //check response
// 		})
// 	}
// }

func TestUpdateUserAPI(t *testing.T) {
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
		PID           string
		body          map[string]interface{}
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			body: map[string]interface{}{
				"first_name": "test_name",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().UpdateAUser(gomock.Any(), user.ID, gomock.Any()).Times(1).Return(user, nil)
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
			buildStubs: func(ustore *mockdb.MockUserInterface) {
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
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "IO-UTIL-ERROR",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "INVALID_FORMAT",
			PID:  "1",
			body: map[string]interface{}{
				"first_name": true,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "ID_NOT_MATCH",
			PID:  "3",
			body: map[string]interface{}{
				"name": "test_group",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UPDATE_INTERNAL_SERVER_ERROR",
			PID:  "1",
			body: map[string]interface{}{
				"first_name": "test_group",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().UpdateAUser(gomock.Any(), user.ID, gomock.Any()).Times(1).Return(&models.User{}, sql.ErrNoRows)
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
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(ustore)    //passed mock stubs
			testServer := &UserTestServer{
				MUser: ustore,
			}
			server := newUserTestServer(t, testServer) //initilize test server
			recorder := httptest.NewRecorder()
			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)
			var request *http.Request
			url := fmt.Sprintf("/user/%v", tc.PID)
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

func TestUpdateProfileAPI(t *testing.T) {
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
		buildStubs    func(ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"first_name": "test_name",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().UpdateAUser(gomock.Any(), user.ID, gomock.Any()).Times(1).Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UNAUTHORIZED",
			body: map[string]interface{}{},
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "IO-UTIL-ERROR",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "INVALID_FORMAT",
			body: map[string]interface{}{
				"first_name": true,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "UPDATE_INTERNAL_SERVER_ERROR",
			body: map[string]interface{}{
				"first_name": "test_group",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().UpdateAUser(gomock.Any(), user.ID, gomock.Any()).Times(1).Return(&models.User{}, sql.ErrNoRows)
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
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(ustore)    //passed mock stubs
			testServer := &UserTestServer{
				MUser: ustore,
			}
			server := newUserTestServer(t, testServer) //initilize test server
			recorder := httptest.NewRecorder()
			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)
			var request *http.Request
			url := "/profile"
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

// func TestGetUsersAPI(t *testing.T) {
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
// 	users := &[]models.User{
// 		{
// 			FirstName:     "test1",
// 			LastName:      "test2",
// 			Image:         "test.jpg",
// 			Company:       "testcomp",
// 			Designation:   "test",
// 			Password:      "test123",
// 			Email:         "bish@gmail.com",
// 			EmailVerified: true,
// 			Active:        true,
// 		},
// 	}
// 	type Query struct {
// 		query         string
// 		page          int
// 		size          int
// 		sortColunm    string
// 		sortDirection string
// 	}
// 	testCases := []struct {
// 		name          string
// 		query         Query
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(ustore *mockdb.MockUserInterface)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "OK",
// 			query: Query{
// 				query:         "test",
// 				page:          1,
// 				size:          5,
// 				sortColunm:    "id",
// 				sortDirection: "desc",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindAllUsersWithFilters(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(users, 1, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				require.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 		{
// 			name:  "UNAUTHORIZED",
// 			query: Query{},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "INTERNAL_SERVER_ERROR",
// 			query: Query{
// 				query:         "test",
// 				page:          1,
// 				size:          5,
// 				sortColunm:    "id",
// 				sortDirection: "desc",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindAllUsersWithFilters(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, 0, sql.ErrConnDone)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				require.Equal(t, http.StatusInternalServerError, recorder.Code)
// 			},
// 		},
// 		{
// 			name:  "NO_QUERY",
// 			query: Query{},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindAllUsersWithFilters(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(users, 1, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				require.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 	}

// 	for i := range testCases {
// 		tc := testCases[i]

// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			tc.setMiddleware(ustore)
// 			tc.buildStubs(ustore)
// 			testServer := &UserTestServer{
// 				MUser: ustore,
// 			}
// 			server := newUserTestServer(t, testServer)
// 			recorder := httptest.NewRecorder()

// 			url := "/users"
// 			request, err := http.NewRequest(http.MethodGet, url, nil)
// 			// request.Header.Add("x-user-role", "ADMIN")
// 			require.NoError(t, err)

// 			// Add query parameters to request URL
// 			q := request.URL.Query()
// 			q.Add("page", fmt.Sprintf("%d", tc.query.page))
// 			q.Add("size", fmt.Sprintf("%d", tc.query.size))
// 			q.Add("sort-direction", fmt.Sprintf(tc.query.sortDirection))
// 			q.Add("search", fmt.Sprintf(tc.query.query))
// 			q.Add("sort-column", fmt.Sprintf(tc.query.sortColunm))

// 			request.URL.RawQuery = q.Encode()
// 			tc.setupAuth(t, request)
// 			server.Router.ServeHTTP(recorder, request)
// 			tc.checkResponse(recorder)
// 		})
// 	}
// }

func TestDeleteUserAPI(t *testing.T) {
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
		buildStubs    func(ustore *mockdb.MockUserInterface)
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
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().DeleteAUser(gomock.Any(), uint(user.ID)).Times(1).Return(int64(1), nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "BAD-REQUEST",
			PID:  "A",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:          "UNAUTHORIZED",
			PID:           "1",
			setupAuth:     func(t *testing.T, request *http.Request) {},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {},
			buildStubs:    func(ustore *mockdb.MockUserInterface) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UNAUTHORIZED",
			PID:  "3",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "INTERNAL_SERVER_ERROR",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().DeleteAUser(gomock.Any(), uint(user.ID)).Times(1).Return(int64(0), sql.ErrNoRows)
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
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(ustore)    //passed mock stubs
			testServer := &UserTestServer{
				MUser: ustore,
			}
			server := newUserTestServer(t, testServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/user/%v", tc.PID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, request) //passed authorization token
			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder) //check response
		})
	}
}

func TestChangeAdminStatusAPI(t *testing.T) {
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
		isAdmin       string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		// {
		// 	name:    "OK",
		// 	PID:     "1",
		// 	isAdmin: "true",
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
		// 	},
		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
		// 	},
		// 	buildStubs: func(ustore *mockdb.MockUserInterface) {
		// 		ustore.EXPECT().DeleteAUser(gomock.Any(), uint(user.ID)).Times(1).Return(int64(1), nil)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusNoContent, recorder.Code)
		// 	},
		// },
		{
			name:    "BAD-REQUEST",
			PID:     "A",
			isAdmin: "false",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:          "UNAUTHORIZED",
			PID:           "1",
			setupAuth:     func(t *testing.T, request *http.Request) {},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {},
			buildStubs:    func(ustore *mockdb.MockUserInterface) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:    "PARSE_ERROR",
			PID:     "1",
			isAdmin: "HELLO",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name:    "USER_NOT_FOUND",
			PID:     "1",
			isAdmin: "true",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(nil, sql.ErrNoRows)
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
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(ustore)    //passed mock stubs
			testServer := &UserTestServer{
				MUser: ustore,
			}
			server := newUserTestServer(t, testServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/user/%v/change-admin-status", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			err = request.ParseForm()
			if err != nil {
				logrus.Error(err)
			}
			request.Form.Set("is_admin", tc.isAdmin)
			tc.setupAuth(t, request) //passed authorization token
			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder) //check response
		})
	}
}

func TestBlockUnBlockAccountAPI(t *testing.T) {
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
		Type          string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "BAD-REQUEST",
			PID:  "A",
			Type: "block",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:          "UNAUTHORIZED",
			PID:           "1",
			Type:          "unblock",
			setupAuth:     func(t *testing.T, request *http.Request) {},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {},
			buildStubs:    func(ustore *mockdb.MockUserInterface) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "EMPTY_TYPE",
			PID:  "1",
			Type: "",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "USER_NOT_FOUND",
			PID:  "1",
			Type: "block",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(nil, sql.ErrNoRows)
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
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(ustore)    //passed mock stubs
			testServer := &UserTestServer{
				MUser: ustore,
			}
			server := newUserTestServer(t, testServer)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/user/%v/block", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			err = request.ParseForm()
			if err != nil {
				logrus.Error(err)
			}
			request.Form.Set("type", tc.Type)
			tc.setupAuth(t, request) //passed authorization token
			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder) //check response
		})
	}
}

func TestDeactivateAccountAPI(t *testing.T) {
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
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name:          "UNAUTHORIZED",
			setupAuth:     func(t *testing.T, request *http.Request) {},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {},
			buildStubs:    func(ustore *mockdb.MockUserInterface) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "USER_NOT_FOUND",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(nil, sql.ErrNoRows)
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
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(ustore)    //passed mock stubs
			testServer := &UserTestServer{
				MUser: ustore,
			}
			server := newUserTestServer(t, testServer)
			recorder := httptest.NewRecorder()
			url := "/user/account/deactivate"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request) //passed authorization token
			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder) //check response
		})
	}
}
