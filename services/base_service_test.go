/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package services

import (
	"log"
	"kcaitech.com/kcserver/utils/sliceutil"
	"testing"
)

func TestGenerateSelectArgs(t *testing.T) {
	result := GenerateSelectArgs(&AccessRecordAndFavoritesQueryResItem{}, "")
	result1 := sliceutil.MapT(func(item *SelectArgs) SelectArgs {
		return *item
	}, *result...)
	log.Println(result1)
}

func TestGenerateJoinArgs(t *testing.T) {
	result := make([]DocumentQueryResItem, 0)
	result1 := GenerateJoinArgs(&result, "document", ParamArgs{"#user_id": "document.user_id"})
	result2 := sliceutil.MapT(func(item *JoinArgs) JoinArgs {
		return *item
	}, *result1...)
	log.Println(result2)
}
