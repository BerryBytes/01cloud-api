package models

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestValidateDNS(t *testing.T) {
	testcases := []struct {
		Name   string
		Dns    *DNS
		cvalue string
		Err    error
	}{
		{
			Name: "AWS_OK",
			Dns: &DNS{
				Name:       "test",
				Provider:   "aws",
				BaseDomain: "test.com.",
				AccessKey:  "TTTTTTEEE",
				SecretKey:  "SECRET",
				Region:     "us-test-1",
				ProjectId:  "cloudflare",
			},
			Err: nil,
		},
		{
			Name: "EMPTY_NAME",
			Dns:  &DNS{},
			Err:  errors.New("required name"),
		},
		{
			Name: "EMPTY_PROVIDER",
			Dns: &DNS{
				Name: "test",
			},
			Err: errors.New("required provider"),
		},
		{
			Name: "EMPTY_BASEDOMAIN",
			Dns: &DNS{
				Name:     "test",
				Provider: "gcp",
			},
			Err: errors.New("required base domain"),
		},
		{
			Name: "BASEDOMAIN_BUT_NO_SUFFRIX",
			Dns: &DNS{
				Name:       "test",
				Provider:   "gcp",
				BaseDomain: "test",
			},
			Err: errors.New("base domain should end with . "),
		},
		{
			Name: "GCP_EMPRTY_CREDENTIAL",
			Dns: &DNS{
				Name:       "test",
				Provider:   "gcp",
				BaseDomain: "test.com.",
			},
			Err: errors.New("required credential file"),
		},
		{
			Name: "GCP_EMPTY_PROJECT_ID",
			Dns: &DNS{
				Name:       "test",
				Provider:   "gcp",
				BaseDomain: "test.com.",
				Credential: "testgcp.json",
			},
			Err: errors.New("required project id"),
		},
		{
			Name: "AWS_EMPRTY_ACCESSKEY",
			Dns: &DNS{
				Name:       "test",
				Provider:   "aws",
				BaseDomain: "test.com.",
			},
			Err: errors.New("required access key"),
		},
		{
			Name: "AWS_EMPRTY_SECRETKEY",
			Dns: &DNS{
				Name:       "test",
				Provider:   "aws",
				BaseDomain: "test.com.",
				AccessKey:  "TTTTTTEEE",
			},
			Err: errors.New("required secret key"),
		},
		{
			Name: "AWS_EMPRTY_REGION",
			Dns: &DNS{
				Name:       "test",
				Provider:   "aws",
				BaseDomain: "test.com.",
				AccessKey:  "TTTTTTEEE",
				SecretKey:  "SECRET",
			},
			Err: errors.New("required aws region"),
		},
		{
			Name: "AWS_NOT_CLOUDFAFE",
			Dns: &DNS{
				Name:       "test",
				Provider:   "aws",
				BaseDomain: "test.com.",
				AccessKey:  "TTTTTTEEE",
				SecretKey:  "SECRET",
				Region:     "us-test-1",
				ProjectId:  "test",
			},
			Err: errors.New("required zone id"),
		},
	}
	for _, tc := range testcases {
		err := tc.Dns.Validate()
		assert.Equal(t, err, tc.Err)
	}

}

func TestValidateCredentials(t *testing.T) {
	testcases := []struct {
		Name   string
		Dns    *DNS
		cvalue string
		Err    error
	}{
		{
			Name: "OK",
			Dns: &DNS{
				Provider:   "gcp",
				Credential: "test.json",
			},
			Err: nil,
		},
		{
			Name: "EMPTY_PROVIDER",
			Dns:  &DNS{},
			Err:  errors.New("required provider name"),
		},
		{
			Name: "GCP_NO_CREDENTIAL",
			Dns: &DNS{
				Provider: "gcp",
			},
			Err: errors.New("required credentials file"),
		},
		{
			Name: "AWS_NO_KEY",
			Dns: &DNS{
				Provider: "aws",
			},
			Err: errors.New("required access key and secret key and region"),
		},
	}
	for _, tc := range testcases {
		err := tc.Dns.ValidateCredentials()
		assert.Equal(t, err, tc.Err)
	}

}

// func TestSaveDNS(t *testing.T) {
// 	test := &DNS{
// 		Name:           "domainhem",
// 		Provider:       "cloudflare",
// 		ProjectId:      "hereshem@gmail.com",
// 		Credential:     "3882f4e7732364a31b2e5b224a05fc5b4c2d9",
// 		TLS:            "zerone-tls-cert",
// 		BaseDomain:     "hem.xyz.np.",
// 		Active:         true,
// 		OrganizationID: 2,
// 	}
// 	test.ID = 1

// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewDNS()
// 	_, err := data.Save(server.DB, test)
// 	if err != nil {
// 		t.Errorf("this is the error creating DNS: %v\n", err)
// 		return
// 	}
// 	//assert.Equal(t, saved.Name, test.Name)
// 	//assert.Equal(t, saved.Provider, test.Provider)
// }

func TestFindDNS(t *testing.T) {
	test := &DNS{
		Name:           "domainhem",
		Provider:       "cloudflare",
		ProjectId:      "hereshem@gmail.com",
		Credential:     "3882f4e7732364a31b2e5b224a05fc5b4c2d9",
		TLS:            "zerone-tls-cert",
		BaseDomain:     "hem.xyz.np.",
		Active:         true,
		OrganizationID: 2,
	}
	test.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "provider"}).AddRow(1, time.Now(), time.Now(), "testdns", "testaws"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewDNS()
	fpc, err := data.Find(server.DB, uint64(test.ID))
	if err != nil {
		t.Errorf("this is the error getting one DNS: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

func TestFindAllDNS(t *testing.T) {
	var testdns *[]DNS
	var err error
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "provider", "region"}).AddRow(1, "testname", "cloudfare", "testregion"))

	data := NewDNS()
	testdns, err = data.FindAll(server.DB)
	if err != nil {
		t.Errorf("this is the error getting the DNS: %v\n", err)
		return
	}
	assert.Equal(t, len(*testdns), 1)
}

func TestFindAllByOrganizationDNS(t *testing.T) {
	var testdns *[]DNS
	var err error
	OrgID := 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "provider", "region"}).AddRow(1, "testname", "cloudfare", "testregion"))

	data := NewDNS()
	testdns, err = data.FindAllByOrganization(server.DB, uint(OrgID))
	if err != nil {
		t.Errorf("this is the error getting the DNS: %v\n", err)
		return
	}
	assert.Equal(t, len(*testdns), 1)
}

func TestFindAllWithInactiveDNS(t *testing.T) {
	var testdns *[]DNS
	var err error
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "provider", "region"}).AddRow(1, "testname", "cloudfare", "testregion"))

	data := NewDNS()
	testdns, err = data.FindAllWithInactive(server.DB)
	if err != nil {
		t.Errorf("this is the error getting the DNS: %v\n", err)
		return
	}
	assert.Equal(t, len(*testdns), 1)
}

func TestUpdateDNS(t *testing.T) {
	test := &DNS{
		Name:           "domainhem",
		Provider:       "cloudflare",
		ProjectId:      "hereshem@gmail.com",
		Credential:     "3882f4e7732364a31b2e5b224a05fc5b4c2d9",
		TLS:            "zerone-tls-cert",
		BaseDomain:     "hem.xyz.np.",
		Active:         true,
		OrganizationID: 2,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "provider", "region"}).AddRow(1, test.Name, test.Provider, test.Region))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "provider", "region"}).AddRow(1, "test", "cloudfare", "us-east-1"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	test.ID = 1
	data := NewDNS()
	_, err := data.Update(server.DB, test)
	if err != nil {
		t.Errorf("this is the error updating the DNS: %v\n", err)
		return
	}
	//assert.Equal(t, updatedactivity.Name, test.Name)
	//assert.Equal(t, updatedactivity.Module, activity.Module)
}

func TestDeleteDNS(t *testing.T) {
	test := &DNS{
		Name:           "domainhem",
		Provider:       "cloudflare",
		ProjectId:      "hereshem@gmail.com",
		Credential:     "3882f4e7732364a31b2e5b224a05fc5b4c2d9",
		TLS:            "zerone-tls-cert",
		BaseDomain:     "hem.xyz.np.",
		Active:         true,
		OrganizationID: 2,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "provider", "region"}).AddRow(1, "testname", "cloudfare", "testregion"))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewDNS()
	id, err := data.Delete(server.DB, uint64(test.ID))
	if err != nil {
		t.Errorf("this is the error deleting the dns: %v\n", err)
		return
	}
	assert.Equal(t, int64(id), int64(1))
}
