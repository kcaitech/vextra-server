/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */


console.log('开始运行所有测试...');

// 导入所有测试文件
import './comment';
import './share';
import './team';
import './users';


// 由于每个测试文件都有自己的 runAllTests 函数，
// 所以不需要在这里调用任何函数。
// 导入测试文件后，它们会自动执行。
