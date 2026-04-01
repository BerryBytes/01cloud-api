package models

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestChartBeforeCreate(t *testing.T) {
	testCase := struct {
		name string
		data *Chart
	}{
		name: "Chart_Test",
		data: &Chart{
			ID: "1",
		},
	}
	tc := testCase
	t.Run(tc.name, func(t *testing.T) {
		err := tc.data.BeforeCreate(server.DB)
		if err != nil {
			logrus.Error(err)
		}
	})
}
