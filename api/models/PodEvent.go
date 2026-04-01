package models

import (
	v1 "k8s.io/api/core/v1"
	"time"
)

type PodEvent struct {
	Namespace string       `json:"namespace"`
	PodName   string       `json:"pod_name"`
	StartedAt time.Time    `json:"started_at"`
	Status    string       `json:"status"`
	Log       string       `json:"log"`
	PodStatus v1.PodStatus `json:"pod_status"`
}
