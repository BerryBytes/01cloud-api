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

type UserInviteTServer struct {
	MUser       models.UserInterface
	MUserInvite models.UserInviteInterface
	MOrg        models.OrganizationInterface
	Router      *mux.Router
}

func NewUserInviteTServer(t *testing.T, store *UserInviteTServer) *Server {
	server, err := newUserInviteServer(store)
	require.NoError(t, err)
	return server
}

func newUserInviteServer(store *UserInviteTServer) (*Server, error) {
	server := &Server{
		MockInterface: store,
		Router:        mux.NewRouter(),
		Cache:         cache.NewCache(1 * time.Minute),
	}
	if gserver, ok := server.MockInterface.(*UserInviteTServer); ok {
		userinviteInterface = gserver.MUserInvite
		userInterface = gserver.MUser
		orgInterface = gserver.MOrg
		middlewares.UserI = gserver.MUser
	}
	server.initializeRoutes("http://api.example.com")
	return server, nil
}

// func TestCreateUserInviteAPI(t *testing.T) {
// 	user := &models.User{
// 		FirstName:     "test1",
// 		LastName:      "test2",
// 		Image:         "test.jpg",
// 		Company:       "testcomp",
// 		Designation:   "test",
// 		Password:      "test123",
// 		Email:         "user@example.com",
// 		EmailVerified: true,
// 		Active:        true,
// 		IsAdmin:       true,
// 	}
// 	user.ID = 1
// 	org := &models.Organization{
// 		Name: "testorg",
// 	}
// 	org.ID = 1

// 	UserInvite := &models.UserInvite{
// 		FirstName: "test",
// 		Email:     "user@example.com",
// 	}
// 	UserInvite.ID = 1
// 	// UserInvites := &[]models.UserInvite{
// 	// 	{
// 	// 		Name: "test",
// 	// 	},
// 	// }
// 	testCases := []struct {
// 		name          string
// 		body          map[string]interface{}
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(dstore *mockdb.MockUserInviteInterface, ustore *mockdb.MockUserInterface)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "OK",
// 			body: map[string]interface{}{
// 				"first_name": "test",
// 				"email":      "user@example.com",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				//	ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockUserInviteInterface, ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), UserInvite.Email).Times(1).Return(&models.User{}, sql.ErrNoRows)
// 				gstore.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(UserInvite, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},
// 		// {
// 		// 	name: "UNAUTHORIZED",
// 		// 	body: map[string]interface{}{},
// 		// 	setupAuth: func(t *testing.T, request *http.Request) {
// 		// 	},
// 		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 		// 	},
// 		// 	buildStubs: func(gstore *mockdb.MockUserInviteInterface, ustore *mockdb.MockUserInterface) {
// 		// 	},
// 		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
// 		// 		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
// 		// 	},
// 		// },
// 		{
// 			name: "IO-UTIL-ERROR",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(gstore *mockdb.MockUserInviteInterface, ustore *mockdb.MockUserInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNPROCESSABLE-ENITITY",
// 			body: map[string]interface{}{
// 				"first_name": 12345,
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(gstore *mockdb.MockUserInviteInterface, ustore *mockdb.MockUserInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "VALIDATION",
// 			body: map[string]interface{}{
// 				"first_name": "bish",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(gstore *mockdb.MockUserInviteInterface, ustore *mockdb.MockUserInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		// {
// 		// 	name: "VALIDATION_DATA",
// 		// 	body: map[string]interface{}{
// 		// 		"name": "aa",
// 		// 	},
// 		// 	setupAuth: func(t *testing.T, request *http.Request) {
// 		// 		addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 		// 	},
// 		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 		// 		ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 		// 	},
// 		// 	buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
// 		// 	},
// 		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
// 		// 		assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 		// 	},
// 		// },
// 		{
// 			name: "FIND_INTERNAL_SERVER_ERROR",
// 			body: map[string]interface{}{
// 				"first_name": "test",
// 				"email":      "user@example.com",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(gstore *mockdb.MockUserInviteInterface, ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), UserInvite.Email).Times(1).Return(user, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "SAVE_INTERNAL_SERVER_ERROR",
// 			body: map[string]interface{}{
// 				"first_name": "test",
// 				"email":      "user@example.com",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(gstore *mockdb.MockUserInviteInterface, ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), UserInvite.Email).Times(1).Return(&models.User{}, sql.ErrNoRows)

// 				gstore.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(&models.UserInvite{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},
// 	}
// 	for i := range testCases {
// 		tc := testCases[i]

// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
// 			gstore := mockdb.NewMockUserInviteInterface(ctrl)
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			tc.setMiddleware(ustore)      //passed middleware
// 			tc.buildStubs(gstore, ustore) //passed mock stubs
// 			tUserInviteserver := &UserInviteTServer{
// 				MUserInvite: gstore,
// 				MUser:       ustore,
// 			}
// 			server := NewUserInviteTServer(t, tUserInviteserver) //initilize test server
// 			recorder := httptest.NewRecorder()
// 			// Marshal body data to JSON
// 			data, err := json.Marshal(tc.body)
// 			assert.NoError(t, err)
// 			var request *http.Request
// 			url := "/user-invite-request"
// 			if tc.name == "IO-UTIL-ERROR" {
// 				request, err = http.NewRequest(http.MethodPost, url, errReader(0))
// 			} else {
// 				request, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
// 			}
// 			assert.NoError(t, err)
// 			tc.setupAuth(t, request)
// 			server.Router.ServeHTTP(recorder, request)
// 			tc.checkResponse(recorder)
// 		})
// 	}
// }

func TestGETUserInviteAPI(t *testing.T) {
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

	UserInvite := &models.UserInvite{
		FirstName: "test",
		Email:     "user@example.com",
		Token:     "asdf",
	}

	UserInvite.ID = 1
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(dstore *mockdb.MockUserInviteInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "asdf",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
			},
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
				gstore.EXPECT().FindByToken(gomock.Any(), UserInvite.Token).Times(1).Return(UserInvite, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			PID:  "asdf",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
			},
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
				gstore.EXPECT().FindByToken(gomock.Any(), UserInvite.Token).Times(1).Return(&models.UserInvite{}, sql.ErrNoRows)

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
			gstore := mockdb.NewMockUserInviteInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(gstore)    //passed mock stubs
			tUserInviteserver := &UserInviteTServer{
				MUserInvite: gstore,
				MUser:       ustore,
			}
			server := NewUserInviteTServer(t, tUserInviteserver)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/user-invite-register/%v", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, request) //passed authorization token

			server.Router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder) //check response
		})
	}
}

func TestGETUserInviteLISTAPI(t *testing.T) {
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

	UserInvite := &[]models.UserInvite{
		{
			FirstName: "test",
			Email:     "user@example.com",
			Token:     "asdf",
		},
	}
	//possible cases
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(dstore *mockdb.MockUserInviteInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)

			},
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
				gstore.EXPECT().FindAll(gomock.Any()).Times(1).Return(UserInvite, nil)
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
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)

			},
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
				gstore.EXPECT().FindAll(gomock.Any()).Times(1).Return(&[]models.UserInvite{}, sql.ErrNoRows)
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
			gstore := mockdb.NewMockUserInviteInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(gstore)    //passed mock stubs
			tUserInviteserver := &UserInviteTServer{
				MUserInvite: gstore,
				MUser:       ustore,
			}
			server := NewUserInviteTServer(t, tUserInviteserver) //initilize test server
			recorder := httptest.NewRecorder()

			url := "/user-invites"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			//request.FormValue(tc.query)
			assert.NoError(t, err)

			tc.setupAuth(t, request) //passed authorization token

			server.Router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder) //check response
		})
	}
}

func TestUpdateteUserInviteAPI(t *testing.T) {
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

	UserInvite := &models.UserInvite{
		FirstName: "test",
		Email:     "user@example.com",
		Token:     "asdf",
	}
	UserInvite.ID = 1
	testCases := []struct {
		name          string
		PID           string
		body          map[string]interface{}
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(dstore *mockdb.MockUserInviteInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			body: map[string]interface{}{
				"first_name": "test",
				"email":      "user@example.com",
				"token":      "asdf",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserInvite.ID)).Times(1).Return(UserInvite, nil)
				gstore.EXPECT().Update(gomock.Any(), gomock.Any()).Times(1).Return(UserInvite, nil)
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
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
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
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
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
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserInvite.ID)).Times(1).Return(UserInvite, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "UserInvite_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserInvite.ID)).Times(1).Return(&models.UserInvite{}, sql.ErrNoRows)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "INVALID_FORMAT",
			PID:  "1",
			body: map[string]interface{}{
				"first_name": true,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserInvite.ID)).Times(1).Return(UserInvite, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "UPDATE_INTERNAL_SERVER_ERROR",
			PID:  "1",
			body: map[string]interface{}{
				"first_name": "test",
				"email":      "user@example.com",
				"token":      "asdf",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserInvite.ID)).Times(1).Return(UserInvite, nil)
				gstore.EXPECT().Update(gomock.Any(), gomock.Any()).Times(1).Return(&models.UserInvite{}, sql.ErrNoRows)
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
			gstore := mockdb.NewMockUserInviteInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(gstore)    //passed mock stubs
			tUserInviteserver := &UserInviteTServer{
				MUserInvite: gstore,
				MUser:       ustore,
			}
			server := NewUserInviteTServer(t, tUserInviteserver) //initilize test server
			recorder := httptest.NewRecorder()
			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)
			var request *http.Request
			url := fmt.Sprintf("/user-invite/%v", tc.PID)
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

func TestDELETEUserInviteAPI(t *testing.T) {
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

	UserInvite := &models.UserInvite{
		FirstName: "test",
		Email:     "user@example.com",
		Token:     "asdf",
	}

	UserInvite.ID = 1
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(dstore *mockdb.MockUserInviteInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserInvite.ID)).Times(1).Return(UserInvite, nil)
				gstore.EXPECT().Delete(gomock.Any(), uint64(UserInvite.ID)).Times(1).Return(int64(UserInvite.ID), nil)
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
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
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
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(dstore *mockdb.MockUserInviteInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UserInvite_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserInvite.ID)).Times(1).Return(&models.UserInvite{}, sql.ErrNoRows)

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
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
			},
			buildStubs: func(gstore *mockdb.MockUserInviteInterface) {
				gstore.EXPECT().Find(gomock.Any(), uint64(UserInvite.ID)).Times(1).Return(UserInvite, nil)
				gstore.EXPECT().Delete(gomock.Any(), uint64(UserInvite.ID)).Times(1).Return(int64(0), sql.ErrNoRows)
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
			gstore := mockdb.NewMockUserInviteInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore) //passed middleware
			tc.buildStubs(gstore)    //passed mock stubs
			tUserInviteserver := &UserInviteTServer{
				MUserInvite: gstore,
				MUser:       ustore,
			}
			server := NewUserInviteTServer(t, tUserInviteserver)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/user-invite/%v", tc.PID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, request) //passed authorization token

			server.Router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder) //check response
		})
	}
}
