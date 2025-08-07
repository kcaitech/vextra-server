/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package reflect

import (
	"log"
	"reflect"
	"testing"
)

type StructA struct {
	FieldA string `json:"field_a"`
}

func Test0(t *testing.T) {
	a := &StructA{
		FieldA: "a",
	}
	dataType := EnterPointer(reflect.TypeOf(a))
	field := dataType.Field(0)
	log.Println(field)
}
