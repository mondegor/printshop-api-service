package xtype_test

import (
	"testing"

	"github.com/mondegor/go-core/mrtype"
	"github.com/stretchr/testify/assert"

	"print-shop-back/internal/warehousing/xtype"
)

func TestNewStoreCursor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params mrtype.CursorParams
		want   xtype.StoreCursor
	}{
		{
			name:   "first page",
			params: mrtype.CursorParams{Value: "", Limit: 20},
			want:   xtype.StoreCursor{Limit: 20},
		},
		{
			name:   "value without separator is the first page",
			params: mrtype.CursorParams{Value: "S1-01", Limit: 20},
			want:   xtype.StoreCursor{Limit: 20},
		},
		{
			name:   "territory and code",
			params: mrtype.CursorParams{Value: "7|S1-01", Limit: 20},
			want:   xtype.StoreCursor{TerritoryID: 7, Code: "S1-01", Limit: 20},
		},
		{
			name:   "code with separator",
			params: mrtype.CursorParams{Value: "7|S1|01", Limit: 20},
			want:   xtype.StoreCursor{TerritoryID: 7, Code: "S1|01", Limit: 20},
		},
		{
			name:   "empty code",
			params: mrtype.CursorParams{Value: "7|", Limit: 20},
			want:   xtype.StoreCursor{TerritoryID: 7, Limit: 20},
		},
		{
			name:   "max territory",
			params: mrtype.CursorParams{Value: "9223372036854775807|S1-01", Limit: 20},
			want:   xtype.StoreCursor{TerritoryID: 9223372036854775807, Code: "S1-01", Limit: 20},
		},
		{
			name:   "territory above int8 range is the first page",
			params: mrtype.CursorParams{Value: "9223372036854775808|S1-01", Limit: 20},
			want:   xtype.StoreCursor{Limit: 20},
		},
		{
			name:   "zero territory is the first page",
			params: mrtype.CursorParams{Value: "0|S1-01", Limit: 20},
			want:   xtype.StoreCursor{Limit: 20},
		},
		{
			name:   "negative territory is the first page",
			params: mrtype.CursorParams{Value: "-1|S1-01", Limit: 20},
			want:   xtype.StoreCursor{Limit: 20},
		},
		{
			name:   "unparsable territory is the first page",
			params: mrtype.CursorParams{Value: "abc|S1-01", Limit: 20},
			want:   xtype.StoreCursor{Limit: 20},
		},
		{
			name:   "zero limit",
			params: mrtype.CursorParams{Value: "7|S1-01", Limit: 0},
			want:   xtype.StoreCursor{TerritoryID: 7, Code: "S1-01", Limit: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, xtype.NewStoreCursor(tt.params))
		})
	}
}
