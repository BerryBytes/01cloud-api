package models

import "time"

type Insight struct {
	Namespace string `json:"namespace,omitempty"`
	Start     int64  `json:"start_time"`
	End       int64  `json:"end_time"`
}

func (m Insight) StartTime() time.Time {
	return time.Unix(0, m.Start*int64(time.Second))
}
func (m Insight) EndTime() time.Time {
	return time.Unix(0, m.End*int64(time.Second))
}
