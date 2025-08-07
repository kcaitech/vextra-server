/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package models

// 自己定义一份，防此auth端变更
type UserProfile struct {
	// DefaultModelData
	Nickname string `json:"nickname"` // Nickname
	Avatar   string `json:"avatar"`   // Avatar URL
	Id       string `json:"id"`       // User ID
}

func (user UserProfile) MarshalJSON() ([]byte, error) {

	return MarshalJSON(user)
}
