/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package my_map

import (
	"log"
	"testing"
)

func Test0(t *testing.T) {
	m := NewSyncMap[string, any]()
	log.Println(m)
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	m.Set("d", 4)
	m.Set("e", 5)
	log.Println(m)
	m.Delete("c")
	log.Println(m)
	m.Range(func(key string, value any) bool {
		log.Println(key, value)
		return true
	})
	log.Println(m.Len())
	m.Clear()
	log.Println(m)
	log.Println(m.Len())
}
