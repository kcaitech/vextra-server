/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package handlers

import (
	"testing"

	"kcaitech.com/kcserver/utils/sliceutil"
)

func TestAllowedKeyList(t *testing.T) {
	key := "Preferences"
	filted := sliceutil.FilterT(func(code string) bool {
		return key == code
	}, AllowedKeyList...)

	if len(filted) != 0 {
		t.Fatalf("%d", len(filted))
	}
}
