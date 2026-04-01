package controllers

import (
	"01cloud-api/api/middlewares"
	"01cloud-api/api/models"
	"errors"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type GroupTServer struct {
	MUser  models.UserInterface
	MGroup models.GroupInterface
	MOrg   models.OrganizationInterface
	Router *mux.Router
}

func NewGroupTServer(t *testing.T, store *GroupTServer) *Server {
	server, err := newGroupServer(store)
	require.NoError(t, err)
	return server
}

func newGroupServer(store *GroupTServer) (*Server, error) {
	server := &Server{
		MockInterface: store,
		Router:        mux.NewRouter(),
	}
	if gserver, ok := server.MockInterface.(*GroupTServer); ok {
		grepo = gserver.MGroup
		userInterface = gserver.MUser
		orgInterface = gserver.MOrg
		middlewares.UserI = gserver.MUser
	}
	server.initializeRoutes("http://api.example.com")
	return server, nil
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

// func TestCreateGroupAPI(t *testing.T) {
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
// 		Name: "testorg",
// 	}
// 	org.ID = 1

// 	group := &models.Group{
// 		Name:           "test",
// 		OrganizationID: 1,
// 		Organization:   org,
// 	}
// 	group.ID = 1
// 	groups := &[]models.Group{
// 		{
// 			Name:           "test",
// 			OrganizationID: 1,
// 			Organization:   org,
// 		},
// 	}
// 	testCases := []struct {
// 		name          string
// 		body          map[string]interface{}
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(dstore *mockdb.MockGroupInterface)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "OK",
// 			body: map[string]interface{}{
// 				"name": "test_group",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().FindAll(gomock.Any(), uint64(org.ID), gomock.Any()).Times(1).Return(groups, nil)
// 				gstore.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(group, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusCreated, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNAUTHORIZED",
// 			body: map[string]interface{}{},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "IO-UTIL-ERROR",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNPROCESSABLE-ENITITY",
// 			body: map[string]interface{}{
// 				"name": 12345,
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "VALIDATE_ERROR",
// 			body: map[string]interface{}{
// 				"name": "",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "GROUPS_NOT_FOUND",
// 			body: map[string]interface{}{
// 				"name": "test_group",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().FindAll(gomock.Any(), uint64(org.ID), gomock.Any()).Times(1).Return(&[]models.Group{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "DUPLICATE_NAME_ERROR",
// 			body: map[string]interface{}{
// 				"name": "test",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().FindAll(gomock.Any(), uint64(org.ID), gomock.Any()).Times(1).Return(groups, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "SAVE_INTERNAL_SERVER_ERROR",
// 			body: map[string]interface{}{
// 				"name": "test_group",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().FindAll(gomock.Any(), uint64(org.ID), gomock.Any()).Times(1).Return(groups, nil)
// 				gstore.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(&models.Group{}, sql.ErrNoRows)
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
// 			gstore := mockdb.NewMockGroupInterface(ctrl)
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			tc.setMiddleware(ustore) //passed middleware
// 			tc.buildStubs(gstore)    //passed mock stubs
// 			tgroupserver := &GroupTServer{
// 				MGroup: gstore,
// 				MUser:  ustore,
// 			}
// 			server := NewGroupTServer(t, tgroupserver) //initilize test server
// 			recorder := httptest.NewRecorder()
// 			// Marshal body data to JSON
// 			data, err := json.Marshal(tc.body)
// 			assert.NoError(t, err)
// 			var request *http.Request
// 			url := "/groups"
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

// func TestGETGROUPAPI(t *testing.T) {
// 	user := &models.User{
// 		FirstName: "test1",
// 		LastName:  "test1",
// 		Email:     "test@gmai.com",
// 		Active:    true,
// 		IsAdmin:   true,
// 		Company:   "test",
// 	}
// 	user.ID = 1
// 	org := &models.Organization{
// 		Name: "testorg",
// 	}
// 	org.ID = 1

// 	group := &models.Group{
// 		Name:           "domainhem",
// 		OrganizationID: 1,
// 		Organization:   org,
// 	}

// 	group.ID = 1
// 	testCases := []struct {
// 		name          string
// 		PID           string
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(dstore *mockdb.MockGroupInterface)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "OK",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(group.OrganizationID))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "BAD-REQUEST",
// 			PID:  "A",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(group.OrganizationID))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(dstore *mockdb.MockGroupInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "InternalError",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(group.OrganizationID))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(&models.Group{}, sql.ErrNoRows)

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
// 			gstore := mockdb.NewMockGroupInterface(ctrl)
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			tc.setMiddleware(ustore) //passed middleware
// 			tc.buildStubs(gstore)    //passed mock stubs
// 			tgroupserver := &GroupTServer{
// 				MGroup: gstore,
// 				MUser:  ustore,
// 			}
// 			server := NewGroupTServer(t, tgroupserver)
// 			recorder := httptest.NewRecorder()

// 			url := fmt.Sprintf("/groups/%v", tc.PID)
// 			request, err := http.NewRequest(http.MethodGet, url, nil)
// 			assert.NoError(t, err)

// 			tc.setupAuth(t, request) //passed authorization token

// 			server.Router.ServeHTTP(recorder, request)

// 			tc.checkResponse(recorder) //check response
// 		})
// 	}
// }

// func TestGETGROUPLISTAPI(t *testing.T) {
// 	user := &models.User{
// 		FirstName: "test1",
// 		LastName:  "test1",
// 		Email:     "test@gmai.com",
// 		Active:    true,
// 		IsAdmin:   true,
// 		Company:   "test",
// 	}
// 	user.ID = 1
// 	org := &models.Organization{
// 		Name: "testorg",
// 	}
// 	org.ID = 1

// 	group := &[]models.Group{
// 		{
// 			Name:           "test",
// 			OrganizationID: 1,
// 			Organization:   org,
// 		},
// 	}
// 	//possible cases
// 	testCases := []struct {
// 		name          string
// 		PID           string
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(dstore *mockdb.MockGroupInterface)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "OK",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().FindAll(gomock.Any(), uint64(org.ID), gomock.Any()).Times(1).Return(group, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "InternalError",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().FindAll(gomock.Any(), uint64(org.ID), gomock.Any()).Times(1).Return(&[]models.Group{}, sql.ErrNoRows)
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
// 			gstore := mockdb.NewMockGroupInterface(ctrl)
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			tc.setMiddleware(ustore) //passed middleware
// 			tc.buildStubs(gstore)    //passed mock stubs
// 			tgroupserver := &GroupTServer{
// 				MGroup: gstore,
// 				MUser:  ustore,
// 			}
// 			server := NewGroupTServer(t, tgroupserver) //initilize test server
// 			recorder := httptest.NewRecorder()

// 			url := "/groups"
// 			request, err := http.NewRequest(http.MethodGet, url, nil)
// 			//request.FormValue(tc.query)
// 			assert.NoError(t, err)

// 			tc.setupAuth(t, request) //passed authorization token

// 			server.Router.ServeHTTP(recorder, request)

// 			tc.checkResponse(recorder) //check response
// 		})
// 	}
// }

// func TestUpdateteGroupAPI(t *testing.T) {
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
// 		Name: "testorg",
// 	}
// 	org.ID = 1

// 	group := &models.Group{
// 		Name:           "test",
// 		OrganizationID: 1,
// 		Organization:   org,
// 	}
// 	group.ID = 1
// 	groups := &[]models.Group{
// 		{
// 			Name:           "test",
// 			OrganizationID: 1,
// 			Organization:   org,
// 		},
// 	}
// 	testCases := []struct {
// 		name          string
// 		PID           string
// 		body          map[string]interface{}
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(dstore *mockdb.MockGroupInterface)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "OK",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"name": "test_gr",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 				gstore.EXPECT().FindAll(gomock.Any(), uint64(org.ID), gomock.Any()).Times(1).Return(groups, nil)
// 				gstore.EXPECT().Update(gomock.Any(), gomock.Any()).Times(1).Return(group, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNAUTHORIZED",
// 			PID:  "1",
// 			body: map[string]interface{}{},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNPROCESSABLE-ENITITY",
// 			PID:  "A",
// 			body: map[string]interface{}{},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "GROUP_NOT_FOUND",
// 			PID:  "1",
// 			body: map[string]interface{}{},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(&models.Group{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "IO-UTIL-ERROR",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "INVALID_FORMAT",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"name": true,
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "GROUPS_NOT_FOUND",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"name": "test_group",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 				gstore.EXPECT().FindAll(gomock.Any(), uint64(org.ID), gomock.Any()).Times(1).Return(&[]models.Group{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "DUPLICATE_NAME_ERROR",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"name": "test",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 				gstore.EXPECT().FindAll(gomock.Any(), uint64(org.ID), gomock.Any()).Times(1).Return(groups, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UPDATE_INTERNAL_SERVER_ERROR",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"name": "test_group",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 				gstore.EXPECT().FindAll(gomock.Any(), uint64(org.ID), gomock.Any()).Times(1).Return(groups, nil)
// 				gstore.EXPECT().Update(gomock.Any(), gomock.Any()).Times(1).Return(&models.Group{}, sql.ErrNoRows)
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
// 			gstore := mockdb.NewMockGroupInterface(ctrl)
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			tc.setMiddleware(ustore) //passed middleware
// 			tc.buildStubs(gstore)    //passed mock stubs
// 			tgroupserver := &GroupTServer{
// 				MGroup: gstore,
// 				MUser:  ustore,
// 			}
// 			server := NewGroupTServer(t, tgroupserver) //initilize test server
// 			recorder := httptest.NewRecorder()
// 			// Marshal body data to JSON
// 			data, err := json.Marshal(tc.body)
// 			assert.NoError(t, err)
// 			var request *http.Request
// 			url := fmt.Sprintf("/groups/%v", tc.PID)
// 			if tc.name == "IO-UTIL-ERROR" {
// 				request, err = http.NewRequest(http.MethodPut, url, errReader(0))
// 			} else {
// 				request, err = http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
// 			}
// 			assert.NoError(t, err)
// 			tc.setupAuth(t, request)
// 			server.Router.ServeHTTP(recorder, request)
// 			tc.checkResponse(recorder)
// 		})
// 	}
// }

// func TestDELETEGROUPAPI(t *testing.T) {
// 	user := &models.User{
// 		FirstName: "test1",
// 		LastName:  "test1",
// 		Email:     "test@gmai.com",
// 		Active:    true,
// 		IsAdmin:   true,
// 		Company:   "test",
// 	}
// 	users := []*models.User{
// 		{
// 			FirstName: "test1",
// 			LastName:  "test1",
// 			Email:     "test@gmai.com",
// 			Active:    true,
// 			IsAdmin:   true,
// 			Company:   "test",
// 		},
// 		{
// 			FirstName: "test2",
// 			LastName:  "test2",
// 			Email:     "test2@gmai.com",
// 			Active:    true,
// 			IsAdmin:   true,
// 			Company:   "test1",
// 		},
// 	}
// 	user.ID = 1
// 	org := &models.Organization{
// 		Name: "testorg",
// 	}
// 	org.ID = 1

// 	group := &models.Group{
// 		Name:           "domainhem",
// 		OrganizationID: 1,
// 		Organization:   org,
// 	}
// 	groupm := &models.Group{
// 		Name:           "domainhem",
// 		OrganizationID: 1,
// 		Organization:   org,
// 		Members:        users,
// 	}

// 	group.ID = 1
// 	testCases := []struct {
// 		name          string
// 		PID           string
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(dstore *mockdb.MockGroupInterface)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "OK",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(group.OrganizationID))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 				gstore.EXPECT().Delete(gomock.Any(), int64(group.ID)).Times(1).Return(int64(group.ID), nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusNoContent, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNAUTHORIZED",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "BAD-REQUEST",
// 			PID:  "A",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(group.OrganizationID))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(dstore *mockdb.MockGroupInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "GROUP_NOT_FOUND",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(group.OrganizationID))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(&models.Group{}, sql.ErrNoRows)

// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "MEMBER_EXIST",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(group.OrganizationID))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(groupm, nil)
// 				//gstore.EXPECT().Delete(gomock.Any(), int64(group.ID)).Times(1).Return(int64(0), sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "DELETE_ERROR",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(group.OrganizationID))
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 				gstore.EXPECT().Delete(gomock.Any(), int64(group.ID)).Times(1).Return(int64(0), sql.ErrNoRows)
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
// 			gstore := mockdb.NewMockGroupInterface(ctrl)
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			tc.setMiddleware(ustore) //passed middleware
// 			tc.buildStubs(gstore)    //passed mock stubs
// 			tgroupserver := &GroupTServer{
// 				MGroup: gstore,
// 				MUser:  ustore,
// 			}
// 			server := NewGroupTServer(t, tgroupserver)
// 			recorder := httptest.NewRecorder()

// 			url := fmt.Sprintf("/groups/%v", tc.PID)
// 			request, err := http.NewRequest(http.MethodDelete, url, nil)
// 			assert.NoError(t, err)

// 			tc.setupAuth(t, request) //passed authorization token

// 			server.Router.ServeHTTP(recorder, request)

// 			tc.checkResponse(recorder) //check response
// 		})
// 	}
// }

// func TestAddMemeberToGroupAPI(t *testing.T) {
// 	user := &models.User{
// 		FirstName:     "test1",
// 		LastName:      "test2",
// 		Image:         "test.jpg",
// 		Company:       "testcomp",
// 		Designation:   "test",
// 		Password:      "test123",
// 		Email:         "test@gmail.com",
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

// 	group := &models.Group{
// 		Name:           "test",
// 		OrganizationID: 1,
// 		Organization:   org,
// 	}
// 	group.ID = 1
// 	testCases := []struct {
// 		name          string
// 		PID           string
// 		body          map[string]interface{}
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface, ostore *mockdb.MockOrganizationInterface)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "ADD_MEMBER",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"email": "test@gmail.com",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface, ostore *mockdb.MockOrganizationInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), gomock.Any()).Times(1).Return(user, nil)
// 				ostore.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(org, nil)
// 				gstore.EXPECT().AddMember(gomock.Any(), user, group).Times(1).Return(nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNAUTHORIZED",
// 			PID:  "1",
// 			body: map[string]interface{}{},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface, ostore *mockdb.MockOrganizationInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNPROCESSABLE-ENITITY",
// 			PID:  "A",
// 			body: map[string]interface{}{},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface, ostore *mockdb.MockOrganizationInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "GROUP_NOT_FOUND",
// 			PID:  "1",
// 			body: map[string]interface{}{},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface, ostore *mockdb.MockOrganizationInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(&models.Group{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "IO-UTIL-ERROR",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface, ostore *mockdb.MockOrganizationInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "INVALID_FORMAT",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"email": true,
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface, ostore *mockdb.MockOrganizationInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "USER_NOT_FOUND",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"email": "test@gmail.com",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface, ostore *mockdb.MockOrganizationInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), gomock.Any()).Times(1).Return(&models.User{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "ORG_NOT_FOUND",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"email": "test@gmail.com",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface, ostore *mockdb.MockOrganizationInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), gomock.Any()).Times(1).Return(user, nil)
// 				ostore.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(&models.Organization{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "ADD_MEMBER_ERROR",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"email": "test@gmail.com",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface, ostore *mockdb.MockOrganizationInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), gomock.Any()).Times(1).Return(user, nil)
// 				ostore.EXPECT().Find(gomock.Any(), gomock.Any()).Times(1).Return(org, nil)
// 				gstore.EXPECT().AddMember(gomock.Any(), user, group).Times(1).Return(errors.New("error occured"))
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
// 			gstore := mockdb.NewMockGroupInterface(ctrl)
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			ostore := mockdb.NewMockOrganizationInterface(ctrl)
// 			tc.setMiddleware(ustore)              //passed middleware
// 			tc.buildStubs(gstore, ustore, ostore) //passed mock stubs
// 			tgroupserver := &GroupTServer{
// 				MGroup: gstore,
// 				MUser:  ustore,
// 				MOrg:   ostore,
// 			}
// 			server := NewGroupTServer(t, tgroupserver) //initilize test server
// 			recorder := httptest.NewRecorder()
// 			// Marshal body data to JSON
// 			data, err := json.Marshal(tc.body)
// 			assert.NoError(t, err)
// 			var request *http.Request
// 			url := fmt.Sprintf("/groups/%v/members", tc.PID)
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

// func TestDeleteMemeberFromGroupAPI(t *testing.T) {
// 	user := &models.User{
// 		FirstName:     "test1",
// 		LastName:      "test2",
// 		Image:         "test.jpg",
// 		Company:       "testcomp",
// 		Designation:   "test",
// 		Password:      "test123",
// 		Email:         "test@gmail.com",
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

// 	group := &models.Group{
// 		Name:           "test",
// 		OrganizationID: 1,
// 		Organization:   org,
// 	}
// 	group.ID = 1
// 	testCases := []struct {
// 		name          string
// 		PID           string
// 		body          map[string]interface{}
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "DELETE_MEMBER",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"email": "test@gmail.com",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), gomock.Any()).Times(1).Return(user, nil)
// 				gstore.EXPECT().DeleteMember(gomock.Any(), user, group).Times(1).Return(nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusNoContent, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNAUTHORIZED",
// 			PID:  "1",
// 			body: map[string]interface{}{},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNPROCESSABLE-ENITITY",
// 			PID:  "A",
// 			body: map[string]interface{}{},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "GROUP_NOT_FOUND",
// 			PID:  "1",
// 			body: map[string]interface{}{},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(&models.Group{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "IO-UTIL-ERROR",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "INVALID_FORMAT",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"email": true,
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "USER_NOT_FOUND",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"email": "test@gmail.com",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), gomock.Any()).Times(1).Return(&models.User{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "DELETE_MEMBER_ERROR",
// 			PID:  "1",
// 			body: map[string]interface{}{
// 				"email": "test@gmail.com",
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(org.ID),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(gstore *mockdb.MockGroupInterface, ustore *mockdb.MockUserInterface) {
// 				gstore.EXPECT().Find(gomock.Any(), uint64(group.ID), uint(group.OrganizationID)).Times(1).Return(group, nil)
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), gomock.Any()).Times(1).Return(user, nil)
// 				gstore.EXPECT().DeleteMember(gomock.Any(), user, group).Times(1).Return(errors.New("error occured"))
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
// 			gstore := mockdb.NewMockGroupInterface(ctrl)
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			tc.setMiddleware(ustore)      //passed middleware
// 			tc.buildStubs(gstore, ustore) //passed mock stubs
// 			tgroupserver := &GroupTServer{
// 				MGroup: gstore,
// 				MUser:  ustore,
// 			}
// 			server := NewGroupTServer(t, tgroupserver) //initilize test server
// 			recorder := httptest.NewRecorder()
// 			// Marshal body data to JSON
// 			data, err := json.Marshal(tc.body)
// 			assert.NoError(t, err)
// 			var request *http.Request
// 			url := fmt.Sprintf("/groups/%v/members", tc.PID)
// 			if tc.name == "IO-UTIL-ERROR" {
// 				request, err = http.NewRequest(http.MethodDelete, url, errReader(0))
// 			} else {
// 				request, err = http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
// 			}
// 			assert.NoError(t, err)
// 			tc.setupAuth(t, request)
// 			server.Router.ServeHTTP(recorder, request)
// 			tc.checkResponse(recorder)
// 		})
// 	}
// }
