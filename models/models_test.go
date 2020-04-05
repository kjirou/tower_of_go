package models

import (
	"github.com/kjirou/tower_of_go/utils"
	"testing"
)

func TestField(t *testing.T) {
	t.Run("At", func(t *testing.T) {
		field := createField(2, 3)

		t.Run("指定した位置の要素を取得できる", func(t *testing.T) {
			var position utils.IMatrixPosition = &utils.MatrixPosition{Y: 1, X: 2}
			element, err := field.At(position)
			if err != nil {
				t.Error(err)
			} else if element.GetPosition().GetY() != 1 {
				t.Error("Y が違う")
			} else if element.GetPosition().GetX() != 2 {
				t.Error("X が違う")
			}
		})
	})
}
