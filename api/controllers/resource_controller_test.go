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

type ResourceTestServer struct {
	MUser         models.UserInterface
	Mresource     models.ResourceInterface
	MOrganization models.OrganizationInterface
	Router        *mux.Router
}

func newResourceTestServer(t *testing.T, store *ResourceTestServer) *Server {
	server, err := NewResourceServer(store)
	require.NoError(t, err)
	return server
}

func NewResourceServer(store *ResourceTestServer) (*Server, error) {
	server := &Server{
		MockInterface: store,
		Router:        mux.NewRouter(),
		Cache:         cache.NewCache(time.Minute * 1),
	}
	if pserver, ok := server.MockInterface.(*ResourceTestServer); ok {
		resourceInterface = pserver.Mresource
		userInterface = pserver.MUser
		orgInterface = pserver.MOrganization
		middlewares.UserI = pserver.MUser

	}
	server.initializeRoutes("http://api.example.com")
	return server, nil
}

func TestCreateResourceAPI(t *testing.T) {
	org := &models.Organization{
		Name: "Basic Plan",
	}
	org.ID = 1
	tResource := &models.Resource{
		Name:       "test-pro",
		Cores:      555,
		Memory:     256,
		Attributes: "kkk",
	}
	tResource.ID = 1
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
		buildStubs    func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		// {
		// 	name: "OK",
		// 	body: map[string]interface{}{
		// 		"name":       "test-pro",
		// 		"cores":      555,
		// 		"memory":     256,
		// 		"attributes": "kkk",
		// 	},
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
		// 	},
		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
		// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
		// 	},
		// 	buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
		// 		// umock := mockdb.NewMockUserInterface(gomock.NewController(t))
		// 		ustore.EXPECT().FindUserByID(gomock.Any(), user.ID).Times(1).Return(user, nil)
		// 		dstore.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(tResource, nil)

		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusCreated, recorder.Code)
		// 	},
		// },
		{
			name: "UNPROCESSABLE-ENITITY",
			body: map[string]interface{}{
				"name":   12345,
				"cores":  555,
				"memory": 256,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		// {
		// 	name: "SAVE_INTERNAL_SERVER_ERROR",
		// 	body: map[string]interface{}{
		// 		"name":   "test_resource",
		// 		"cores":  555,
		// 		"memory": 256,
		// 	},
		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
		// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
		// 	},
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
		// 	},
		// 	buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
		// 		ustore.EXPECT().FindUserByID(gomock.Any(), user.ID).Times(1).Return(user, nil)
		// 		dstore.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(&models.Resource{}, sql.ErrNoRows)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		// 	},
		// },

		{
			name: "IO-UTIL-ERROR",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},

		// {
		// 	name: "SAVEBYID_INTERNAL_SERVER_ERROR",
		// 	body: map[string]interface{}{
		// 		"name": "test_resource",
		// 	},
		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
		// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
		// 	},
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
		// 	},
		// 	buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
		// 		ustore.EXPECT().FindUserByID(gomock.Any(), user.ID).Times(1).Return(&models.User{}, sql.ErrConnDone)
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
			pstore := mockdb.NewMockResourceInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			sstore := mockdb.NewMockOrganizationInterface(ctrl)
			tc.buildStubs(pstore, ustore, sstore)
			tc.setMiddleware(ustore)
			tResourceServer := &ResourceTestServer{
				MUser:         ustore,
				Mresource:     pstore,
				MOrganization: sstore,
			}
			pserver := newResourceTestServer(t, tResourceServer)
			recorder := httptest.NewRecorder()
			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)
			var request *http.Request
			url := "/resource"
			if tc.name == "IO-UTIL-ERROR" {
				request, err = http.NewRequest(http.MethodPost, url, errReader(0))
			} else {
				request, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			}
			assert.NoError(t, err)
			tc.setupAuth(t, request)
			pserver.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetResourceAPI(t *testing.T) {
	org := &models.Organization{
		Name: "Basic Plan",
	}
	org.ID = 1
	tResource := &models.Resource{
		Name:       "test-pro",
		Attributes: "kkk",
	}
	tResource.ID = 1
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
		setMiddleware func(ustore *mockdb.MockUserInterface)
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(dstore *mockdb.MockResourceInterface)
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
			buildStubs: func(dstore *mockdb.MockResourceInterface) {
				dstore.EXPECT().Find(gomock.Any(), uint64(tResource.ID)).Times(1).Return(tResource, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name: "NotFoundID",
			PID:  "rrr",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockResourceInterface) {
				store.EXPECT().
					Find(gomock.Any(), gomock.Any()).
					Times(0).
					Return(&models.Resource{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockResourceInterface) {
				store.EXPECT().
					Find(gomock.Any(), gomock.Eq(uint64(tResource.ID))).
					Times(1).
					Return(&models.Resource{}, sql.ErrConnDone)
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
			pstore := mockdb.NewMockResourceInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(pstore)

			tResourceServer := &ResourceTestServer{
				Mresource: pstore,
				MUser:     ustore,
			}
			pserver := newResourceTestServer(t, tResourceServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/resource/%v", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)
			pserver.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGETRESOURCELISTAPI(t *testing.T) {
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

	resource := &[]models.Resource{
		{
			Name:           "test",
			OrganizationID: 1,
			Organization:   org,
		},
	}
	//possible cases
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), user.ID).Times(1).Return(user, nil)
				gstore.EXPECT().FindAll(gomock.Any(), gomock.Any()).Times(1).Return(resource, nil)
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
			buildStubs: func(gstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "SAVEBYID_INTERNAL_SERVER_ERROR",
			PID:  "1",
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			buildStubs: func(gstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), user.ID).Times(1).Return(&models.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "InternalError",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(org.ID), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(gstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)

				gstore.EXPECT().FindAll(gomock.Any(), gomock.Any()).Times(1).Return(&[]models.Resource{}, sql.ErrNoRows)
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
			gstore := mockdb.NewMockResourceInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore)      //passed middleware
			tc.buildStubs(gstore, ustore) //passed mock stubs
			tresourceserver := &ResourceTestServer{
				Mresource: gstore,
				MUser:     ustore,
			}
			server := newResourceTestServer(t, tresourceserver) //initilize test server
			recorder := httptest.NewRecorder()

			url := "/resources"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			//request.FormValue(tc.query)
			assert.NoError(t, err)

			tc.setupAuth(t, request) //passed authorization token

			server.Router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder) //check response
		})
	}
}

func TestAdminResourceAPI(t *testing.T) {
	org := &models.Organization{
		Name: "Basic Plan",
	}
	org.ID = 1
	resource := &[]models.Resource{
		{
			Name:           "test",
			OrganizationID: 1,
			Organization:   org,
		},
	}
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
	testuser := &models.User{
		FirstName:     "test1",
		LastName:      "test2",
		Image:         "test.jpg",
		Company:       "testcomp",
		Designation:   "test",
		Password:      "test123",
		Email:         "bish@gmail.com",
		EmailVerified: true,
		Active:        true,
		IsAdmin:       false,
	}
	testuser.ID = 1
	testCases := []struct {
		name          string
		body          map[string]interface{}
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), gomock.Eq(user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), gomock.Eq(uint(user.ID))).Times(1).Return(user, nil)
				dstore.EXPECT().FindAllWithInactive(gomock.Any(), gomock.Any()).Times(1).Return(resource, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},

		{
			name: "SAVEBYID_INTERNAL_SERVER_ERROR",
			body: map[string]interface{}{
				"name": "test_resource",
			},
			PID: "1",
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), user.ID).Times(1).Return(&models.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "Unauthorized_ERROR",
			body: map[string]interface{}{
				"name": "test_resource",
			},
			PID: "1",
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), testuser.ID).Times(1).Return(testuser, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "SAVE_INTERNAL_SERVER_ERROR",
			PID:  "1",
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), user.ID).Times(1).Return(user, nil)
				dstore.EXPECT().FindAllWithInactive(gomock.Any(), gomock.Any()).Times(1).Return(&[]models.Resource{}, sql.ErrConnDone)
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
			pstore := mockdb.NewMockResourceInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			sstore := mockdb.NewMockOrganizationInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(pstore, ustore, sstore)
			tResourceServer := &ResourceTestServer{
				MUser:         ustore,
				Mresource:     pstore,
				MOrganization: sstore,
			}
			pserver := newResourceTestServer(t, tResourceServer)
			recorder := httptest.NewRecorder()
			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)
			var request *http.Request
			url := "/admin/resources"
			request, err = http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
			assert.NoError(t, err)
			tc.setupAuth(t, request)
			pserver.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateResourceAPI(t *testing.T) {
	org := &models.Organization{
		Name: "Basic Plan",
	}
	org.ID = 1
	tResource := &models.Resource{
		Name:       "test-pro",
		Attributes: "kkk",
	}
	tResource.ID = 1
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
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			body: map[string]interface{}{
				"name":       "test-pro",
				"attributes": "kkk",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), gomock.Eq(user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), gomock.Eq(uint(user.ID))).Times(1).Return(user, nil)
				dstore.EXPECT().Find(gomock.Any(), uint64(tResource.ID)).Times(1).Return(tResource, nil)
				dstore.EXPECT().Update(gomock.Any(), gomock.Any()).Times(1).Return(tResource, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UNPROCESSABLE-ENITITY",
			body: map[string]interface{}{
				"name": 12345,
			},
			PID: "A",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "SAVEBYID_INTERNAL_SERVER_ERROR",
			body: map[string]interface{}{
				"name": "test_resource",
			},
			PID: "1",
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), user.ID).Times(1).Return(&models.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "SAVE_INTERNAL_SERVER_ERROR",
			body: map[string]interface{}{
				"name": "test_resource",
			},
			PID: "1",
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), user.ID).Times(1).Return(user, nil)
				dstore.EXPECT().Find(gomock.Any(), gomock.Eq(uint64(tResource.ID))).Times(1).Return(&models.Resource{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},

		// {
		// 	name: "UPDATE_INTERNAL_SERVER_ERROR",
		// 	body: map[string]interface{}{
		// 		"name": "test_resource",
		// 	},
		// 	PID: "1",
		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
		// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
		// 	},
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
		// 	},
		// 	buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
		// 		ustore.EXPECT().FindUserByID(gomock.Any(), user.ID).Times(1).Return(user, nil)
		// 		dstore.EXPECT().Find(gomock.Any(), uint64(tResource.ID)).Times(1).Return(tResource, nil)
		// 		dstore.EXPECT().Update(gomock.Any(), gomock.Any()).Times(1).Return(&models.Resource{}, sql.ErrNoRows)

		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		// 	},
		// },
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
			buildStubs: func(dstore *mockdb.MockResourceInterface, ustore *mockdb.MockUserInterface, sstore *mockdb.MockOrganizationInterface) {
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
			pstore := mockdb.NewMockResourceInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			sstore := mockdb.NewMockOrganizationInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(pstore, ustore, sstore)
			tResourceServer := &ResourceTestServer{
				MUser:         ustore,
				Mresource:     pstore,
				MOrganization: sstore,
			}
			pserver := newResourceTestServer(t, tResourceServer)
			recorder := httptest.NewRecorder()
			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)
			var request *http.Request
			url := fmt.Sprintf("/resource/%v", tc.PID)
			if tc.name == "IO-UTIL-ERROR" {
				request, err = http.NewRequest(http.MethodPut, url, errReader(0))
			} else {
				request, err = http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			}
			assert.NoError(t, err)
			tc.setupAuth(t, request)
			pserver.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
