package controllers

import (
	mockdb "01cloud-api/api/controllers/mocks"
	"01cloud-api/api/models"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetByIDActivityAPI(t *testing.T) {
	activity := &models.Activity{
		Remarks: "good",
		Action:  "aa",
	}
	activity.ID = 1
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
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(store *mockdb.MockActivityInterface)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockActivityInterface) {
				store.EXPECT().Find(gomock.Any(), gomock.Eq(uint64(activity.ID))).Times(1).Return(activity, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, 200)
			},
		},
		{
			name: "NotFoundID",
			PID:  "rrr",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockActivityInterface) {
				store.EXPECT().Find(gomock.Any(), gomock.Any()).Times(0).Return(&models.Activity{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "InternalError",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockActivityInterface) {
				store.EXPECT().Find(gomock.Any(), gomock.Eq(uint64(activity.ID))).Times(1).Return(&models.Activity{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			astore := mockdb.NewMockActivityInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(astore)
			store := &ATServer{
				MActivity: astore,
				MUser:     ustore,
			}
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/activity/%v", tc.PID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)
			server.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateActivityAPI(t *testing.T) {
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
	activity := &models.Activity{
		Action:    "aa",
		Module:    "xxz",
		Active:    true,
		Remarks:   "Nepal",
		ProjectID: 1,
	}
	activity.ID = 1
	testCases := []struct {
		name          string
		body          map[string]interface{}
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(store *mockdb.MockActivityInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"action":     "aa",
				"module":     "xxz",
				"active":     true,
				"remarks":    "Nepal",
				"project_id": 1,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockActivityInterface) {
				store.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Times(1).
					Return(activity, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name: "IO-UTIL-ERROR",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockActivityInterface) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: map[string]interface{}{
				"action":     "",
				"module":     "xxz",
				"active":     true,
				"remarks":    "Nepal",
				"project_id": 1,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockActivityInterface) {
				store.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0).Return(&models.Activity{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "UnprocessableEnitity",
			body: map[string]interface{}{
				"action":     "sss",
				"module":     "xxz",
				"active":     true,
				"remarks":    "",
				"project_id": 1,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockActivityInterface) {
				store.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(&models.Activity{}, sql.ErrNoRows)
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
			astore := mockdb.NewMockActivityInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(astore)
			store := &ATServer{
				MActivity: astore,
				MUser:     ustore,
			}
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)
			var request *http.Request
			url := "/activity"
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

func TestDeleteActivityAPI(t *testing.T) {
	activity := &models.Activity{
		Action:    "aa",
		Module:    "xxz",
		Active:    true,
		Remarks:   "Nepal",
		ProjectID: 1,
	}
	activity.ID = 1
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
		setupAuth     func(t *testing.T, request *http.Request)
		setMiddleware func(ustore *mockdb.MockUserInterface)
		buildStubs    func(store *mockdb.MockActivityInterface)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  "1",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockActivityInterface) {
				store.EXPECT().Find(gomock.Any(), gomock.Eq(uint64(activity.ID))).Times(1).Return(activity, nil)
				store.EXPECT().Delete(gomock.Any(), gomock.Eq(uint64(activity.ID))).Times(1).Return(int64(1), nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NotFound",
			PID:  "rrr",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockActivityInterface) {
				store.EXPECT().Find(gomock.Any(), gomock.Any()).Times(0).Return(&models.Activity{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			astore := mockdb.NewMockActivityInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(astore)
			store := &ATServer{
				MActivity: astore,
				MUser:     ustore,
			}
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/activity/%v", tc.PID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)
			server.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})
	}
}

// func TestGetActivitiesAPI(t *testing.T) {
// 	activities := &[]models.Activity{
// 		{
// 			Remarks: "good",
// 			Action:  "aa",
// 		},
// 	}
// 	//activity.ID = 1
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
// 	testCases := []struct {
// 		name          string
// 		PID           string
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		setMiddleware func(ustore *mockdb.MockUserInterface)
// 		buildStubs    func(store *mockdb.MockActivityInterface)
// 		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "OK",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(1),user.Email)
// 			},
// 			setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 				ustore.EXPECT().FindUserByEmail(gomock.Any(),(user.Ema)).Times(1).Return(user, nil)
// 			},
// 			buildStubs: func(store *mockdb.MockActivityInterface) {
// 				store.EXPECT().FindAll(gomock.Any()).Times(1).Return(activities, nil)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusOK, 200)
// 			},
// 		},
// 		// {
// 		// 	name: "NotFoundID",
// 		// 	PID:  "rrr",
// 		// 	setupAuth: func(t *testing.T, request *http.Request) {
// 		// 		addAuthorization(t, request, uint(user.ID), uint(1),user.Email)
// 		// 	},
// 		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(),(user.Ema)).Times(1).Return(user, nil)
// 		// 	},
// 		// 	buildStubs: func(store *mockdb.MockActivityInterface) {
// 		// 		store.EXPECT().Find(gomock.Any(), gomock.Any()).Times(0).Return(&models.Activity{}, nil)
// 		// 	},
// 		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 		// 		assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
// 		// 	},
// 		// },
// 		// {
// 		// 	name: "InternalError",
// 		// 	PID:  "1",
// 		// 	setupAuth: func(t *testing.T, request *http.Request) {
// 		// 		addAuthorization(t, request, uint(user.ID), uint(1),user.Email)
// 		// 	},
// 		// 	setMiddleware: func(ustore *mockdb.MockUserInterface) {
// 		// 		ustore.EXPECT().FindUserByEmail(gomock.Any(),(user.Ema)).Times(1).Return(user, nil)
// 		// 	},
// 		// 	buildStubs: func(store *mockdb.MockActivityInterface) {
// 		// 		store.EXPECT().Find(gomock.Any(), gomock.Eq(uint64(activity.ID))).Times(1).Return(&models.Activity{}, sql.ErrConnDone)
// 		// 	},
// 		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 		// 		assert.Equal(t, http.StatusNotFound, recorder.Code)
// 		// 	},
// 		// },
// 	}
// 	for i := range testCases {
// 		tc := testCases[i]
// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
// 			astore := mockdb.NewMockActivityInterface(ctrl)
// 			ustore := mockdb.NewMockUserInterface(ctrl)
// 			tc.setMiddleware(ustore)
// 			tc.buildStubs(astore)
// 			store := &ATServer{
// 				MActivity: astore,
// 				MUser:     ustore,
// 			}
// 			server := newTestServer(t, store)
// 			recorder := httptest.NewRecorder()

// 			url := "/activities"
// 			request, err := http.NewRequest(http.MethodGet, url, nil)
// 			assert.NoError(t, err)
// 			tc.setupAuth(t, request)
// 			server.Router.ServeHTTP(recorder, request)
// 			//check response
// 			tc.checkResponse(t, recorder)
// 		})
// 	}
// }

func TestUpdateActivityAPI(t *testing.T) {
	activity := models.Activity{
		Action:    "aa",
		Module:    "xxz",
		Active:    true,
		Remarks:   "Nepal",
		ProjectID: 1,
	}
	activity.ID = 1
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
		buildStubs    func(store *mockdb.MockActivityInterface)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			PID:  strconv.FormatUint(uint64(activity.ID), 10),
			body: map[string]interface{}{
				"action":     "aa",
				"module":     "xxz",
				"active":     true,
				"remarks":    "Nepal",
				"project_id": 1,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockActivityInterface) {
				store.EXPECT().Find(gomock.Any(), gomock.Eq(uint64(activity.ID))).Times(1).Return(&activity, nil)
				store.EXPECT().Update(gomock.Any(), activity).Times(1).Return(&activity, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NotFound",
			PID:  "rrr",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockActivityInterface) {
				store.EXPECT().
					Find(gomock.Any(), gomock.Any()).
					Times(0).
					Return(&models.Activity{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "InternalError",
			PID:  strconv.FormatUint(uint64(activity.ID), 10),
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockActivityInterface) {
				store.EXPECT().
					Find(gomock.Any(), gomock.Eq(uint64(activity.ID))).
					Times(1).
					Return(&models.Activity{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			PID:  strconv.FormatUint(uint64(activity.ID), 10),
			body: map[string]interface{}{
				"action":     "aa",
				"module":     "xxz",
				"active":     true,
				"remarks":    "Nepal",
				"project_id": 1,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(1), user.Email)
			},
			setMiddleware: func(ustore *mockdb.MockUserInterface) {
				ustore.EXPECT().FindUserByEmail(gomock.Any(), (user.Email)).Times(1).Return(user, nil)
				ustore.EXPECT().GetSessionById(gomock.Any(), gomock.Any())
			},
			buildStubs: func(store *mockdb.MockActivityInterface) {
				store.EXPECT().Find(gomock.Any(), gomock.Eq(uint64(activity.ID))).Times(1).Return(&activity, nil)
				store.EXPECT().Update(gomock.Any(), activity).Times(1).Return(&models.Activity{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			astore := mockdb.NewMockActivityInterface(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.setMiddleware(ustore)
			tc.buildStubs(astore)
			store := &ATServer{
				MActivity: astore,
				MUser:     ustore,
			}
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/activity/%v", tc.PID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)
			tc.setupAuth(t, request)
			server.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})
	}
}
