package controllers

import (
	"01cloud-api/api/auth"
	mockdb "01cloud-api/api/controllers/mocks"
	"01cloud-api/api/models"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DNSTestServer struct {
	MUser  models.UserInterface
	MDns   models.IDNS
	Router *mux.Router
}

func newDNSTestServer(t *testing.T, store *DNSTestServer) *Server {
	server, err := NewDServer(store)
	require.NoError(t, err)
	return server
}

func NewDServer(store *DNSTestServer) (*Server, error) {
	server := &Server{
		MockInterface: store,
		Router:        mux.NewRouter(),
	}
	server.DNSTestRoutes()
	return server, nil
}

func (server *Server) DNSTestRoutes() {
	if dserver, ok := server.MockInterface.(*DNSTestServer); ok {
		idns = dserver.MDns
		userInterface = dserver.MUser
	}
	server.setJson("/dns/{id}", server.GetDns, "GET")
	server.setJson("/dns", server.CreateDns, "POST")
	server.setJson("/dns", server.GetDnsList, "GET")
}

func TestCreateDNSAPI(t *testing.T) {
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
	dns := &models.DNS{
		Name:           "domainhem",
		Provider:       "cloudflare",
		ProjectId:      "hereshem@gmail.com",
		Credential:     "3882f4e7732364a31b2e5b224a05fc5b4c2d9",
		TLS:            "zerone-tls-cert",
		BaseDomain:     "hem.xyz.np.",
		Active:         true,
		OrganizationID: 2,
	}
	dns.ID = 1
	testCases := []struct {
		name          string
		body          map[string]interface{}
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(dstore *mockdb.MockIDNS, ustore *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		// {
		// 	name: "OK",
		// 	body: map[string]interface{}{
		// 		"name":            "domainhem2",
		// 		"provider":        "cloudflare",
		// 		"project_id":      "hereshem@gmail.com",
		// 		"credentials":     "3882f4e7732364a31b2e5b224a05fc5b4c2d9",
		// 		"tls":             "zerone-tls-cert",
		// 		"base_domain":     "hem.xyz.np.",
		// 		"active":          true,
		// 		"organization_id": 2,
		// 	},
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(dns.OrganizationID),user.Email)
		// 	},
		// 	buildStubs: func(dstore *mockdb.MockIDNS) {
		// 		umock := mockdb.NewMockUserInterface(gomock.NewController(t))
		// 		umock.EXPECT().FindUserByID(gomock.Any(), gomock.Eq(1)).Times(1).Return(user, nil)

		// 		dstore.EXPECT().
		// 			Save(gomock.Any(), gomock.Any()).
		// 			Times(1).
		// 			Return(dns, nil)

		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusOK, recorder.Code)
		// 	},
		{
			name: "UNPROCESSABLE-ENITITY",
			body: map[string]interface{}{
				"name":            "test",
				"provider":        "cloudflare",
				"project_id":      "hereshem@gmail.com",
				"credentials":     "3882f4e7732364a31b2e5b224a05fc5b4c2d9",
				"tls":             "zerone-tls-cert",
				"base_domain":     "hem.xyz.np.",
				"active":          "true",
				"organization_id": 2,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(dns.OrganizationID), user.Email)
			},
			buildStubs: func(dstore *mockdb.MockIDNS, ustore *mockdb.MockUserInterface) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "VALIDATE_ERROR",
			body: map[string]interface{}{
				"provider":        "gcp",
				"project_id":      "hereshem@gmail.com",
				"credentials":     "3882f4e7732364a31b2e5b224a05fc5b4c2d9.json",
				"tls":             "zerone-tls-cert",
				"base_domain":     "hem.xyz.np.",
				"active":          true,
				"organization_id": 2,
				"zone_id":         "test",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(dns.OrganizationID), user.Email)
			},
			buildStubs: func(dstore *mockdb.MockIDNS, ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "UNAUTHORIZED",
			body: map[string]interface{}{
				"name":            "test",
				"provider":        "aws",
				"project_id":      "cloudflare",
				"access_key":      "3882f4e7732364a31b2e5b224a05fc5b4c2d9",
				"secret_key":      "TEST",
				"region":          "us-east-1",
				"base_domain":     "test.com.",
				"active":          true,
				"organization_id": 2,
			},
			setupAuth: func(t *testing.T, request *http.Request) {

			},
			buildStubs: func(dstore *mockdb.MockIDNS, ustore *mockdb.MockUserInterface) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "USER-NOT_FOUND",
			body: map[string]interface{}{
				"name":            "test",
				"provider":        "aws",
				"project_id":      "cloudflare",
				"access_key":      "3882f4e7732364a31b2e5b224a05fc5b4c2d9",
				"secret_key":      "TEST",
				"region":          "us-east-1",
				"base_domain":     "test.com.",
				"active":          true,
				"organization_id": 2,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(dns.OrganizationID), user.Email)
			},
			buildStubs: func(dstore *mockdb.MockIDNS, ustore *mockdb.MockUserInterface) {
				dstore.EXPECT().IsNameExists(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(false)
				ustore.EXPECT().FindUserByID(gomock.Any(), gomock.Eq(uint(user.ID))).Times(1).Return(&models.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
			},
		},
		{
			name: "UNAUTHORIZED-USER-ORG",
			body: map[string]interface{}{
				"name":            "test",
				"provider":        "aws",
				"project_id":      "cloudflare",
				"access_key":      "3882f4e7732364a31b2e5b224a05fc5b4c2d9",
				"secret_key":      "TEST",
				"region":          "us-east-1",
				"base_domain":     "test.com.",
				"active":          true,
				"organization_id": 2,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(0), user.Email)
			},
			buildStubs: func(dstore *mockdb.MockIDNS, ustore *mockdb.MockUserInterface) {
				user.IsAdmin = false
				dstore.EXPECT().IsNameExists(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(false)
				ustore.EXPECT().FindUserByID(gomock.Any(), gomock.Eq(uint(user.ID))).Times(1).Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			dstore := mockdb.NewMockIDNS(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.buildStubs(dstore, ustore)
			tdnsserver := &DNSTestServer{
				MUser: ustore,
				MDns:  dstore,
			}
			server := newDNSTestServer(t, tdnsserver)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)

			url := "/dns"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			assert.NoError(t, err)
			tc.setupAuth(t, request)
			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func addAuthorization(t *testing.T, request *http.Request, uid uint, orgid uint, email string) {
	token, err := auth.CreateToken(uid, orgid, 0)
	require.NoError(t, err)
	authorizationHeader := fmt.Sprintf("%s %s", "basic", token)
	request.Header.Set("Authorization", authorizationHeader)
	request.Header.Set("X-Email", email)
}

// func TestGETDNSAPI(t *testing.T) {
// 	dns := &models.DNS{
// 		Name:           "domainhem",
// 		Provider:       "cloudflare",
// 		ProjectId:      "hereshem@gmail.com",
// 		Credential:     "3882f4e7732364a31b2e5b224a05fc5b4c2d9",
// 		TLS:            "zerone-tls-cert",
// 		BaseDomain:     "hem.xyz.np.",
// 		Active:         true,
// 		OrganizationID: 2,
// 	}
// 	user := &models.User{
// 		FirstName: "test1",
// 		LastName:  "test1",
// 		Email:     "test@gmai.com",
// 		Active:    true,
// 		IsAdmin:   true,
// 		Company:   "test",
// 	}
// 	user.ID = 1
// 	dns.ID = 1
// 	testCases := []struct {
// 		name          string
// 		PID           string
// 		setupAuth     func(t *testing.T, request *http.Request)
// 		buildStubs    func(dstore *mockdb.MockIDNS)
// 		checkResponse func(recoder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "OK",
// 			PID:  "1",
// 			setupAuth: func(t *testing.T, request *http.Request) {
// 				addAuthorization(t, request, uint(user.ID), uint(dns.OrganizationID),user.Email)
// 			},
// 			buildStubs: func(dstore *mockdb.MockIDNS) {
// 				var muser *mockdb.MockUserInterface
// 				ctrl := gomock.NewController(t)
// 				data := mockdb.NewMockUserInterface(ctrl)
// 				muser = data
// 				muser.EXPECT().
// 					FindUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
// 					Times(1).
// 					Return(user, nil)

// 				dstore.EXPECT().
// 					Find(gomock.Any(), gomock.Eq(uint64(dns.ID))).
// 					Times(1).
// 					Return(dns, nil)
// 			},
// 			checkResponse: func(recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 		// {
// 		// 	name: "UNAUTHORIZED",
// 		// 	PID:  "1",
// 		// 	setupAuth: func(t *testing.T, request *http.Request) {
// 		// 	},
// 		// 	buildStubs: func(dstore *mockdb.MockIDNS) {
// 		// 	},
// 		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
// 		// 		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
// 		// 	},
// 		// },

// 		// {
// 		// 	name: "BAD-REQUEST",
// 		// 	PID:  "A",
// 		// 	setupAuth: func(t *testing.T, request *http.Request) {
// 		// 		addAuthorization(t, request, uint(user.ID), uint(dns.OrganizationID),user.Email)
// 		// 	},
// 		// 	buildStubs: func(dstore *mockdb.MockIDNS) {
// 		// 	},
// 		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
// 		// 		assert.Equal(t, http.StatusBadRequest, recorder.Code)
// 		// 	},
// 		// },
// 		// {
// 		// 	name: "InternalError",
// 		// 	PID:  "1",
// 		// 	setupAuth: func(t *testing.T, request *http.Request) {
// 		// 		addAuthorization(t, request, uint(user.ID), uint(dns.OrganizationID),user.Email)
// 		// 	},
// 		// 	buildStubs: func(store *mockdb.MockIDNS) {
// 		// 		store.EXPECT().
// 		// 			Find(gomock.Any(), gomock.Eq(uint64(dns.ID))).
// 		// 			Times(1).
// 		// 			Return(&models.DNS{}, sql.ErrConnDone)
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
// 			dstore := mockdb.NewMockIDNS(ctrl)
// 			tc.buildStubs(dstore)
// 			tdnsserver := &DNSTestServer{
// 				MDns: dstore,
// 			}
// 			server := newDNSTestServer(t, tdnsserver)
// 			recorder := httptest.NewRecorder()

// 			url := fmt.Sprintf("/dns/%v", tc.PID)
// 			request, err := http.NewRequest(http.MethodGet, url, nil)
// 			assert.NoError(t, err)
// 			tc.setupAuth(t, request)

// 			server.Router.ServeHTTP(recorder, request)
// 			//check response
// 			tc.checkResponse(recorder)
// 		})
// 	}
// }

func TestGETDNSLISTAPI(t *testing.T) {
	orgId := 2
	dnsList := &[]models.DNS{
		{
			Name:           "domainhem",
			Provider:       "cloudflare",
			ProjectId:      "hereshem@gmail.com",
			Credential:     "3882f4e7732364a31b2e5b224a05fc5b4c2d9",
			TLS:            "zerone-tls-cert",
			BaseDomain:     "hem.xyz.np.",
			Active:         true,
			OrganizationID: 2,
		},
		{
			Name:           "domainhem",
			Provider:       "cloudflare",
			ProjectId:      "hereshem@gmail.com",
			Credential:     "3882f4e7732364a31b2e5b224a05fc5b4c2d9",
			TLS:            "zerone-tls-cert",
			BaseDomain:     "hem.xyz.np.",
			Active:         true,
			OrganizationID: 2,
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
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(dstore *mockdb.MockIDNS, muser *mockdb.MockUserInterface)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(t, request, uint(user.ID), uint(orgId), user.Email)
			},
			buildStubs: func(dstore *mockdb.MockIDNS, muser *mockdb.MockUserInterface) {
				muser.EXPECT().
					FindUserByID(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				dstore.EXPECT().
					FindAllByOrganization(gomock.Any(), gomock.Eq(uint(orgId))).
					Times(1).
					Return(dnsList, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		// {
		// 	name: "UNAUTHORIZED",
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 	},
		// 	buildStubs: func(dstore *mockdb.MockIDNS, muser *mockdb.MockUserInterface) {

		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		// 	},
		// },

		// {
		// 	name: "No-USER-FOUND",
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(orgId))
		// 	},
		// 	buildStubs: func(dstore *mockdb.MockIDNS, muser *mockdb.MockUserInterface) {
		// 		muser.EXPECT().
		// 			FindUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
		// 			Times(1).
		// 			Return(&models.User{}, sql.ErrNoRows)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "INTERNAL-SERVER-ERROR",
		// 	setupAuth: func(t *testing.T, request *http.Request) {
		// 		addAuthorization(t, request, uint(user.ID), uint(orgId))
		// 	},
		// 	buildStubs: func(dstore *mockdb.MockIDNS, muser *mockdb.MockUserInterface) {
		// 		muser.EXPECT().
		// 			FindUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
		// 			Times(1).
		// 			Return(user, nil)

		// 		dstore.EXPECT().
		// 			FindAllByOrganization(gomock.Any(), gomock.Eq(uint(orgId))).
		// 			Times(1).
		// 			Return(&[]models.DNS{}, sql.ErrNoRows)
		// 	},
		// 	checkResponse: func(recorder *httptest.ResponseRecorder) {
		// 		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		// 	},
		// },
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			dstore := mockdb.NewMockIDNS(ctrl)
			ustore := mockdb.NewMockUserInterface(ctrl)
			tc.buildStubs(dstore, ustore)
			tdnsserver := &DNSTestServer{
				MUser: ustore,
				MDns:  dstore,
			}
			server := newDNSTestServer(t, tdnsserver)
			recorder := httptest.NewRecorder()

			url := "/dns"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			tc.setupAuth(t, request)
			server.Router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(recorder)
		})
	}
}
