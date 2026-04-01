package models

import (
	"time"

	"helm.sh/helm/v3/pkg/release"
)

type K8sRelease struct {
	ReleaseId    string        `json:"id"`
	Name         string        `json:"name"`
	NameSpace    string        `json:"namespace"`
	Version      int           `json:"version"`
	Manifest     string        `json:"manifest" datastore:"manifest,noindex"`
	Info         *release.Info `json:"info"`
	Config       string        `datastore:"config,noindex" json:"config"`
	CName        string        `json:"cname"`
	ValuesString string        `datastore:"values,noindex" json:"values"`
}

type RevisionHistory struct {
	Release []Release `datastore:"releases" json:"releases"`
}

type Release struct {
	Name          string    `datastore:"name" json:"name"`
	Namespace     string    `datastore:"namespace" json:"namespace"`
	Description   string    `datastore:"description" json:"description"`
	Status        string    `datastore:"status" json:"status"`
	Version       int       `datastore:"version" json:"version"`
	Tag           string    `datastore:"tag" json:"tag"`
	FirstDeployed time.Time `datstore:"first_deployed" json:"first_deployed"`
	LastDeployed  time.Time `datastore:"last_deployed" json:"last_deployed"`
}
