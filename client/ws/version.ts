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
import { DataTypes, VersionData } from "./types";

export class Version extends ConnectClient {
    private updateHandlerSet = new Set<(data: VersionData) => void>()
    
    constructor(connect: Connect) {
        super(connect, DataTypes.GenerateVersion)
    }

    onMessage(data: any): void {
        this._onUpdated(data as VersionData)
    }
    
    private _onUpdated(versionData: VersionData) {
        // @ts-ignore
        for (const handler of this.updateHandlerSet) handler(versionData);
    }

    public addUpdatedHandler(onUpdated: (versionData: VersionData) => void) {
        this.updateHandlerSet.add(onUpdated)
    }

    public removeUpdatedHandler(onUpdated: (versionData: VersionData) => void) {
        this.updateHandlerSet.delete(onUpdated)
    }
}