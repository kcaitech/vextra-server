/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

import { Connect, ConnectClient } from "./connect";
import { DataTypes } from "./types";

export class Thumbnail extends ConnectClient {
    constructor(connect: Connect) {
        super(connect, DataTypes.Thumbnail)
    }

    onMessage(data: any): void {
    }

    public async upload(name: string, contentType: string, data: ArrayBuffer): Promise<boolean> {
        await this.waitReady()
        return !!this.sendBinary({ name: name, contentType }, data, 1000, 3);
    }
}