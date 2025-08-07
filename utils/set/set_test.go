/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package set

import (
	"testing"
)

func Test0(t *testing.T) {
	set := NewSet(1, 2, 3, 4, 5)
	set.Remove(3)
	if set.Contains(3) {
		t.Error("set.Contains(3) = true, want false")
	}
	if set.Size() != 4 {
		t.Errorf("set.Size() = %d, want 4", set.Size())
	}
	set.Clear()
	if set.Size() != 0 {
		t.Errorf("set.Size() = %d, want 0", set.Size())
	}
}

func Test1(t *testing.T) {
	var (
		a = byte(0x00)
		b = byte(0x02)
	)
	set := NewSet(a, b)
	if !set.Contains(a) {
		t.Error("set.Contains(a) = false, want true")
	}
}
