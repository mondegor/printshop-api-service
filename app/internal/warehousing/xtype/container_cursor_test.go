package xtype_test

import (
	"testing"

	"github.com/mondegor/go-core/mrtype"
	"github.com/stretchr/testify/assert"

	"print-shop-back/internal/warehousing/xtype"
)

func TestNewContainerCursor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params mrtype.CursorParams
		want   xtype.ContainerCursor
	}{
		{
			name:   "first page",
			params: mrtype.CursorParams{Value: "", Limit: 20},
			want:   xtype.ContainerCursor{Limit: 20},
		},
		{
			name:   "value without separator is the first page",
			params: mrtype.CursorParams{Value: "G/ABC", Limit: 20},
			want:   xtype.ContainerCursor{Limit: 20},
		},
		{
			name:   "code and marker",
			params: mrtype.CursorParams{Value: "G/ABC|5", Limit: 20},
			want:   xtype.ContainerCursor{Code: "G/ABC", Marker: 5, Limit: 20},
		},
		{
			name:   "max marker",
			params: mrtype.CursorParams{Value: "G/ABC|32767", Limit: 20},
			want:   xtype.ContainerCursor{Code: "G/ABC", Marker: 32767, Limit: 20},
		},
		{
			name:   "marker above int2 range is zero",
			params: mrtype.CursorParams{Value: "G/ABC|32768", Limit: 20},
			want:   xtype.ContainerCursor{Code: "G/ABC", Limit: 20},
		},
		{
			name:   "negative marker is zero",
			params: mrtype.CursorParams{Value: "G/ABC|-1", Limit: 20},
			want:   xtype.ContainerCursor{Code: "G/ABC", Limit: 20},
		},
		{
			name:   "unparsable marker is zero",
			params: mrtype.CursorParams{Value: "G/ABC|abc", Limit: 20},
			want:   xtype.ContainerCursor{Code: "G/ABC", Limit: 20},
		},
		{
			name:   "zero limit",
			params: mrtype.CursorParams{Value: "G/ABC|5", Limit: 0},
			want:   xtype.ContainerCursor{Code: "G/ABC", Marker: 5, Limit: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, xtype.NewContainerCursor(tt.params))
		})
	}
}
