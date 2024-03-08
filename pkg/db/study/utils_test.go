package study

import (
	"testing"
)

func TestGetTotalPages(t *testing.T) {
	type args struct {
		totalCount int64
		limit      int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "Test 1",
			args: args{
				totalCount: 10,
				limit:      10,
			},
			want: 1,
		},
		{
			name: "Test 2",
			args: args{
				totalCount: 10,
				limit:      5,
			},
			want: 2,
		},
		{
			name: "Test 3",
			args: args{
				totalCount: 10,
				limit:      3,
			},
			want: 4,
		},
		{
			name: "Test 4",
			args: args{
				totalCount: 10,
				limit:      1,
			},
			want: 10,
		},
		{
			name: "Test 5",
			args: args{
				totalCount: 10,
				limit:      0,
			},
			want: 0,
		},
		{
			name: "Test 6",
			args: args{
				totalCount: 0,
				limit:      10,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTotalPages(tt.args.totalCount, tt.args.limit); got != tt.want {
				t.Errorf("getTotalPages() = %v, want %v", got, tt.want)
			}
		})
	}
}
