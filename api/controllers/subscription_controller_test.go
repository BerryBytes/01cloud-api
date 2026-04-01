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
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SubTestServer struct {
	MUser         models.UserInterface
	MSubscription models.SubscriptionInterface
}

func newSubTestServer(t *testing.T, store *SubTestServer) *Server {
	server, err := NewSubTServer(store)
	require.NoError(t, err)
	return server
}

func NewSubTServer(store *SubTestServer) (*Server, error) {
	server := &Server{
		MockInterface: store,
		Router:        mux.NewRouter(),
		Cache:         cache.NewCache(time.Minute * 1),
	}
	if pserver, ok := server.MockInterface.(*SubTestServer); ok {
		//iproject = pserver.MProject
		userInterface = pserver.MUser
		subscriptionInterface = pserver.MSubscription
		//orgInterface = pserver.MOrg
		//activityInterface = pserver.MActivity
		middlewares.UserI = pserver.MUser
	}
	server.initializeRoutes("http://api.example.com")
	return server, nil
}

// func TestCreateSubscriptionAPI(t *testing.T) {
// 	user := &models.User{
// 		FirstName: "test1",
// 		LastName:  "test1",
// 		Email:     "test@gmai.com",
// 		Active:    true,
// 		Company:   "test",
// 	}
// 	user.ID = 1
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
// 	testCases := []struct {
// 		name          string
// 		body          map[string]interface{}
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(pstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		// {
// 		// 	name: "OK",
// 		// 	PID:  "1",
// 		// 	setupAuth: func(t *testing.T, request *http.Request) {
// 		// 		addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
// 		// 	},
// 		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 		// 	},
// 		// 	buildStubs: func(sstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
// 		// 		sstore.EXPECT().Find(gomock.Any(), uint64(1)).Times(1).Return(subscription, nil)
// 		// 	},
// 		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
// 		// 		assert.Equal(t, http.StatusOK, recorder.Code)
// 		// 	},
// 		// },
// 		{
// 			name: "UNAUTHORIZED",
// 			body: map[string]interface{}{},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 			},
// 			buildStubs: func(pstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "IO-UTIL-ERROR",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(1),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(pstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
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
// 				addAuthorization(t, request, uint(user.ID), uint(1),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(pstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
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
// 				addAuthorization(t, request, uint(user.ID), uint(1),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(pstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "USER_NOT_FOUND",
// 			body: map[string]interface{}{
// 				"name":          "test123",
// 				"apps":          10,
// 				"disk_space":    10,
// 				"memory":        2,
// 				"cores":         1,
// 				"data_transfer": 10,
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(1),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(sstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(&models.User{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusNotFound, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UNAUTHORIZED_USER",
// 			body: map[string]interface{}{
// 				"name":          "test123",
// 				"apps":          10,
// 				"disk_space":    10,
// 				"memory":        2,
// 				"cores":         1,
// 				"data_transfer": 10,
// 			},
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(sstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusUnauthorized, 401)
// 			},
// 		},
// 		// {
// 		// 	name: "SUBSCRIPTION_NOT_FOUND",
// 		// 	PID:  "1",
// 		// 	setupAuth: func(t *testing.T, request *http.Request) {
// 		// 		addAuthorization(t, request, uint(user.ID), uint(1),user.Email)
// 		// 	},
// 		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
// 		// 	},
// 		// 	buildStubs: func(sstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
// 		// 		sstore.EXPECT().Find(gomock.Any(), uint64(1)).Times(1).Return(&models.Subscription{}, sql.ErrNoRows)
// 		// 	},
// 		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
// 		// 		assert.Equal(t, http.StatusNotFound, recorder.Code)
// 		// 	},
// 		// },
// 	}
// 	for i := range testCases {
// 		tc := testCases[i]

// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			sstore := mockdb.NewMockSubscriptionInterface(ctrl)
// 			tc.setMiddleware(ustore)
// 			tc.buildStubs(sstore, ustore)
// 			subServer := &SubTestServer{
// 				MUser:         ustore,
// 				MSubscription: sstore,
// 			}
// 			server := newSubTestServer(t, subServer)
// 			recorder := httptest.NewRecorder()
// 			// Marshal body data to JSON
// 			data, err := json.Marshal(tc.body)
// 			assert.NoError(t, err)
// 			var request *http.Request
// 			url := "/subscription"
// 			if tc.name == "IO-UTIL-ERROR" {
// 				request, err = http.NewRequest(http.MethodPost, url, errReader(0))
// 			} else {
// 				request, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
// 			}
// 			assert.NoError(t, err)
// 			tc.setupAuth(t, request)

//				server.Router.ServeHTTP(recorder, request)
//				//check response
//				tc.checkResponse(recorder)
//			})
//		}
//	}
func TestGetSubscriptionAPI(t *testing.T) {
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		IsAdmin:   true,
		Company:   "test",
	}
	user.ID = 1
	subscription := &models.Subscription{
		Name:         "Basic Plan",
		Apps:         5,
		DiskSpace:    5,
		Memory:       2,
		Cores:        1,
		DataTransfer: 5,
		Price:        100,
		Weight:       1,
		Backups:      5,
		Active:       true,
	}
	subscription.ID = 1
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(pstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface)
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
			buildStubs: func(sstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
				sstore.EXPECT().Find(gomock.Any(), uint64(1)).Times(1).Return(subscription, nil)
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
			buildStubs: func(sstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
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
			buildStubs: func(pstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "SUBSCRIPTION_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(sstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
				sstore.EXPECT().Find(gomock.Any(), uint64(1)).Times(1).Return(&models.Subscription{}, sql.ErrNoRows)
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
			sstore := mockdb.NewMockSubscriptionInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(sstore, ustore)
			subServer := &SubTestServer{
				MUser:         ustore,
				MSubscription: sstore,
			}
			server := newSubTestServer(t, subServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/subscription/%v", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			server.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}

func TestGetSubscriptionsAPI(t *testing.T) {
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		IsAdmin:   true,
		Company:   "test",
	}
	user.ID = 1
	subscriptions := &[]models.Subscription{
		{
			Model: gorm.Model{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:         "Basic Plan",
			Apps:         5,
			DiskSpace:    5,
			Memory:       2,
			Cores:        1,
			DataTransfer: 5,
			Price:        100,
			Weight:       1,
			Backups:      5,
			Active:       true,
		},
	}
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(pstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface)
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
			buildStubs: func(sstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
				sstore.EXPECT().FindAll(gomock.Any(), gomock.Any()).Times(1).Return(subscriptions, nil)
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
			buildStubs: func(pstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "SUBSCRIPTION_NOT_FOUND",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(sstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
				sstore.EXPECT().FindAll(gomock.Any(), gomock.Any()).Times(1).Return(&[]models.Subscription{}, sql.ErrNoRows)
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
			sstore := mockdb.NewMockSubscriptionInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(sstore, ustore)
			subServer := &SubTestServer{
				MUser:         ustore,
				MSubscription: sstore,
			}
			server := newSubTestServer(t, subServer)
			recorder := httptest.NewRecorder()

			url := "/subscriptions"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			server.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteSubscriptionAPI(t *testing.T) {
	user := &models.User{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "test@gmai.com",
		Active:    true,
		//IsAdmin:   true,
		Company: "test",
	}
	user.ID = 1
	subscription := &models.Subscription{
		Name:         "Basic Plan",
		Apps:         5,
		DiskSpace:    5,
		Memory:       2,
		Cores:        1,
		DataTransfer: 5,
		Price:        100,
		Weight:       1,
		Backups:      5,
		Active:       true,
	}
	subscription.ID = 1
	testCases := []struct {
		name          string
		PID           string
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(pstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		// {
		// 	name: "OK",
		// 	PID:  "1",
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
		// 	},
		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
		// 	},
		// 	buildStubs: func(sstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
		// 		sstore.EXPECT().Find(gomock.Any(), uint64(1)).Times(1).Return(subscription, nil)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusOK, recorder.Code)
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
			buildStubs: func(sstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
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
			buildStubs: func(pstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
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
			buildStubs: func(sstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByID(gomock.Any(), uint(user.ID)).Times(1).Return(&models.User{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		// {
		// 	name: "UNAUTHORIZED_USER",
		// 	PID:  "1",
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(0),user.Email)
		// 	},
		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
		// 		ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
		// 	},
		// 	buildStubs: func(pstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusUnauthorized, 401)
		// 	},
		// },
		// {
		// 	name: "SUBSCRIPTION_NOT_FOUND",
		// 	PID:  "1",
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(1),user.Email)
		// 	},
		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(), user.Email).Times(1).Return(user, nil)
		// 	},
		// 	buildStubs: func(sstore *mockdb.MockSubscriptionInterface, ustore *mockdb.MockUserInterface) {
		// 		sstore.EXPECT().Find(gomock.Any(), uint64(1)).Times(1).Return(&models.Subscription{}, sql.ErrNoRows)
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
			ustore := mockdb.NewMockUserInterface(ctrl)
			sstore := mockdb.NewMockSubscriptionInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(sstore, ustore)
			subServer := &SubTestServer{
				MUser:         ustore,
				MSubscription: sstore,
			}
			server := newSubTestServer(t, subServer)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/subscription/%v", tc.PID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)

			server.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}
