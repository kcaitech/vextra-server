/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */



export const TEST_API_URL = 'http://localhost:80/api'
export const TEST_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoia2NhaSIsInNlc3Npb25faWQiOiJ0QXd1bTdjbFdIIiwidG9rZW5fdHlwZSI6ImFjY2VzcyIsImlzcyI6ImtjYWl0ZWNoLmNvbSIsImV4cCI6MTc0MzgxNzkzNiwiaWF0IjoxNzQzODEwNzM2fQ.BrjGNEcEOIA00Ny6uyR_OerVzPRGQtJafs27Ul4kjUY'
export const TEST_UNAUTHORIZED: () => void = () => {
    console.log("UNAUTHORIZED")
}