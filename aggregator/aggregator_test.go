package aggregator

import "testing"

func Test_avg(t *testing.T) {
	tests := []struct {
		name    string
		ints    []int
		wantAvg float64
	}{
		{"Case 1: [1,3]", []int{1, 3}, 2},
		{"Case 2: []", []int{}, 0},
		{"Case 2: [0, 100]", []int{0, 100}, 50},
		{"Case 2: [-100, 100]", []int{-100, 100}, 0},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAvg := avg(tt.ints); gotAvg != tt.wantAvg {
				t.Errorf("avg() = %v, want %v", gotAvg, tt.wantAvg)
			}
		})
	}
}
