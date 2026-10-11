package imposition_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/pkg/mrcalc/algo/sheet/imposition"
	"print-shop-back/pkg/mrcalc/enum"
	"print-shop-back/pkg/mrcalc/s2/rect2d"
)

// TODO: ДОБАВИТЬ ТЕСТ MIRROR

func TestImposition_CalcWithAllowRotation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		element  rect2d.Format
		distance rect2d.Format
		out      rect2d.Format
		want     imposition.Output
	}{
		// WARNING: большинство данных совпадает с данными теста CalcTotal
		{
			name:     "test1",
			element:  rect2d.Format{Width: 100, Height: 50},
			distance: rect2d.Format{},
			out:      rect2d.Format{Width: 260, Height: 210},
			want: imposition.Output{
				ContainerFormat: rect2d.Format{Width: 250, Height: 200},
				Fragments: rect2d.Fragments{
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 100, Height: 50},
						Layout:   rect2d.Layout{ByWidth: 2, ByHeight: 4},
						Position: enum.PositionTop,
					},
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 50, Height: 100},
						Layout:   rect2d.Layout{ByWidth: 1, ByHeight: 2},
						Position: enum.PositionOnside,
					},
				},
				RestArea:      4600,
				AllowRotation: true,
			},
		},
		{
			name:     "test1",
			element:  rect2d.Format{Width: 88, Height: 48},
			distance: rect2d.Format{Width: 2, Height: 2},
			out:      rect2d.Format{Width: 435, Height: 300}, // 450x320 - frame 15x20
			want: imposition.Output{
				ContainerFormat: rect2d.Format{Width: 408, Height: 298},
				Fragments: rect2d.Fragments{
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 88, Height: 48},
						Distance: rect2d.Format{Width: 2, Height: 2},
						Layout:   rect2d.Layout{ByWidth: 4, ByHeight: 6},
						Position: enum.PositionTop,
					},
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 48, Height: 88},
						Distance: rect2d.Format{Width: 2, Height: 2},
						Layout:   rect2d.Layout{ByWidth: 1, ByHeight: 3},
						Position: enum.PositionOnside,
					},
				},
				RestArea:      10356,
				AllowRotation: true,
			},
		},
		{
			name:     "test2",
			element:  rect2d.Format{Width: 88, Height: 48},
			distance: rect2d.Format{Width: 2, Height: 2},
			out:      rect2d.Format{Width: 435, Height: 290}, // 450x320 - frame 15x30
			want: imposition.Output{
				ContainerFormat: rect2d.Format{Width: 408, Height: 268},
				Fragments: rect2d.Fragments{
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 88, Height: 48},
						Distance: rect2d.Format{Width: 2, Height: 2},
						Layout:   rect2d.Layout{ByWidth: 4, ByHeight: 5},
						Position: enum.PositionTop,
					},
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 48, Height: 88},
						Distance: rect2d.Format{Width: 2, Height: 2},
						Layout:   rect2d.Layout{ByWidth: 1, ByHeight: 3},
						Position: enum.PositionOnside,
					},
				},
				RestArea:      23966,
				AllowRotation: true,
			},
		},
		{
			name:     "test3",
			element:  rect2d.Format{Width: 48, Height: 88},
			distance: rect2d.Format{Width: 2, Height: 2},
			out:      rect2d.Format{Width: 435, Height: 300}, // 450x320 - frame 15x20
			want: imposition.Output{
				ContainerFormat: rect2d.Format{Width: 398, Height: 268},
				Fragments: rect2d.Fragments{
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 48, Height: 88},
						Distance: rect2d.Format{Width: 2, Height: 2},
						Layout:   rect2d.Layout{ByWidth: 8, ByHeight: 3},
						Position: enum.PositionTop,
					},
				},
				RestArea:      23836,
				AllowRotation: true,
			},
		},
		{
			name:     "test4",
			element:  rect2d.Format{Width: 48, Height: 88},
			distance: rect2d.Format{Width: 2, Height: 2},
			out:      rect2d.Format{Width: 435, Height: 290}, // 450x320 - frame 15x30
			want: imposition.Output{
				ContainerFormat: rect2d.Format{Width: 398, Height: 268},
				Fragments: rect2d.Fragments{
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 48, Height: 88},
						Distance: rect2d.Format{Width: 2, Height: 2},
						Layout:   rect2d.Layout{ByWidth: 8, ByHeight: 3},
						Position: enum.PositionTop,
					},
				},
				RestArea:      19486,
				AllowRotation: true,
			},
		},
		{
			name:     "test5",
			element:  rect2d.Format{Width: 420, Height: 300},
			distance: rect2d.Format{},
			out:      rect2d.Format{Width: 435, Height: 300}, // 450x320 - frame 15x20
			want: imposition.Output{
				ContainerFormat: rect2d.Format{Width: 420, Height: 300},
				Fragments: rect2d.Fragments{
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 420, Height: 300},
						Layout:   rect2d.Layout{ByWidth: 1, ByHeight: 1},
						Position: enum.PositionTop,
					},
				},
				RestArea:      4500,
				AllowRotation: true,
			},
		},
		{
			name:     "test6",
			element:  rect2d.Format{Width: 210, Height: 297},
			distance: rect2d.Format{Width: 2, Height: 2},
			out:      rect2d.Format{Width: 435, Height: 300}, // 450x320 - frame 15x20
			want: imposition.Output{
				ContainerFormat: rect2d.Format{Width: 422, Height: 297},
				Fragments: rect2d.Fragments{
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 210, Height: 297},
						Distance: rect2d.Format{Width: 2, Height: 2},
						Layout:   rect2d.Layout{ByWidth: 2, ByHeight: 1},
						Position: enum.PositionTop,
					},
				},
				RestArea:      5166,
				AllowRotation: true,
			},
		},
		{
			// данный тест не зависит от дистанции между изделиями, т.к. изделие всего одно
			name:     "test7",
			element:  rect2d.Format{Width: 235, Height: 100},
			distance: rect2d.Format{Width: 2, Height: 2},
			out:      rect2d.Format{Width: 235, Height: 100}, // 450x320 - frame 15x20
			want: imposition.Output{
				ContainerFormat: rect2d.Format{Width: 235, Height: 100},
				Fragments: rect2d.Fragments{
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 235, Height: 100},
						Distance: rect2d.Format{Width: 2, Height: 2},
						Layout:   rect2d.Layout{ByWidth: 1, ByHeight: 1},
						Position: enum.PositionTop,
					},
				},
				RestArea:      0,
				AllowRotation: true,
			},
		},
		{
			// height = y = 2x + 1; width = x + y + 3; при x = 100 -> y = 201, height = 201, width = 304
			name:     "test8",
			element:  rect2d.Format{Width: 201, Height: 100},
			distance: rect2d.Format{Width: 3, Height: 1},
			out:      rect2d.Format{Width: 304, Height: 201},
			want: imposition.Output{
				ContainerFormat: rect2d.Format{Width: 304, Height: 201},
				Fragments: rect2d.Fragments{
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 201, Height: 100},
						Distance: rect2d.Format{Width: 3, Height: 1},
						Layout:   rect2d.Layout{ByWidth: 1, ByHeight: 2},
						Position: enum.PositionTop,
					},
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 100, Height: 201},
						Distance: rect2d.Format{Width: 1, Height: 3},
						Layout:   rect2d.Layout{ByWidth: 1, ByHeight: 1},
						Position: enum.PositionOnside,
					},
				},
				RestArea:      0,
				AllowRotation: true,
			},
		},
		{
			name:     "test9",
			element:  rect2d.Format{Width: 201, Height: 100},
			distance: rect2d.Format{Width: 1, Height: 1},
			out:      rect2d.Format{Width: 302, Height: 201},
			want: imposition.Output{
				ContainerFormat: rect2d.Format{Width: 302, Height: 201},
				Fragments: rect2d.Fragments{
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 201, Height: 100},
						Distance: rect2d.Format{Width: 1, Height: 1},
						Layout:   rect2d.Layout{ByWidth: 1, ByHeight: 2},
						Position: enum.PositionTop,
					},
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 100, Height: 201},
						Distance: rect2d.Format{Width: 1, Height: 1},
						Layout:   rect2d.Layout{ByWidth: 1, ByHeight: 1},
						Position: enum.PositionOnside,
					},
				},
				RestArea:      0,
				AllowRotation: true,
			},
		},
		{
			// здесь расстояние между элементами не должно влиять
			name:     "test10",
			element:  rect2d.Format{Width: 301, Height: 201},
			distance: rect2d.Format{Width: 10, Height: 10},
			out:      rect2d.Format{Width: 302, Height: 201},
			want: imposition.Output{
				ContainerFormat: rect2d.Format{Width: 301, Height: 201},
				Fragments: rect2d.Fragments{
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 301, Height: 201},
						Distance: rect2d.Format{Width: 10, Height: 10},
						Layout:   rect2d.Layout{ByWidth: 1, ByHeight: 1},
						Position: enum.PositionTop,
					},
				},
				RestArea:      201,
				AllowRotation: true,
			},
		},
		{
			name:     "test11",
			element:  rect2d.Format{Width: 301, Height: 202},
			distance: rect2d.Format{Width: 10, Height: 10},
			out:      rect2d.Format{Width: 302, Height: 201},
			want: imposition.Output{
				ContainerFormat: rect2d.Format{},
				Fragments:       nil,
				RestArea:        0,
			},
		},
		{
			// bug fixed
			name:     "test12",
			element:  rect2d.Format{Width: 100, Height: 200},
			distance: rect2d.Format{Width: 2, Height: 2},
			out:      rect2d.Format{Width: 442, Height: 304}, // 450x320 - frame 8x16
			want: imposition.Output{
				ContainerFormat: rect2d.Format{Width: 406, Height: 302},
				Fragments: rect2d.Fragments{
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 100, Height: 200},
						Distance: rect2d.Format{Width: 2, Height: 2},
						Layout:   rect2d.Layout{ByWidth: 4, ByHeight: 1},
						Position: enum.PositionTop,
					},
					rect2d.Fragment{
						Element:  rect2d.Format{Width: 200, Height: 100},
						Distance: rect2d.Format{Width: 2, Height: 2},
						Layout:   rect2d.Layout{ByWidth: 2, ByHeight: 1},
						Position: enum.PositionBottom,
					},
				},
				RestArea:      12364,
				AllowRotation: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			im := imposition.New(log.NopLogger())

			got, err := im.Calc(context.Background(), tt.element, tt.distance, tt.out, imposition.Options{AllowRotation: true})
			assert.Equal(t, tt.want, got)
			assert.NoError(t, err)
		})
	}
}
