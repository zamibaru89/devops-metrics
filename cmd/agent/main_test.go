package main

import "testing"

func TestMetricCounter_UpdateMetrics(t *testing.T) {
	type fields struct {
		PollCount uint64
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "positive test #1",
			fields: fields{
				PollCount: 1,
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MetricCounter{
				PollCount: tt.fields.PollCount,
			}
			m.UpdateMetrics()
		})
	}
}
