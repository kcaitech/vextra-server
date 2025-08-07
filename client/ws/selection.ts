/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

import { throttle } from "./timing_util";
import { Connect, ConnectClient } from "./connect";
import { DataTypes, DocSelectionData, DocSelectionOpData, IClientContext, SelectionEvents } from "./types";

export class Selection extends ConnectClient {
    private context: IClientContext;
    private onMessageList: ((data: DocSelectionOpData) => void)[] = [];
    private docSelectionOpUpdate: typeof this._update | undefined;
    private selectionWatcherForOp = this._selectionWatcherForOp.bind(this);

    constructor(connect: Connect, context: IClientContext) {
        super(connect, DataTypes.Selection);
        this.context = context;
        context.selection.watch(this.selectionWatcherForOp);
    }

    onMessage(data: any): void {
        for (const onMessage of this.onMessageList) onMessage(data);
    }

    private async _update(data: DocSelectionData, timeout?: number): Promise<boolean> {
        return !!this.send(data, timeout ?? 0);
    }

    private _selectionWatcherForOp(type: string) {
        if (![SelectionEvents.page_change, SelectionEvents.shape_change, SelectionEvents.shape_hover_change, SelectionEvents.text_change].includes(type as SelectionEvents)) return;
        if (!this.docSelectionOpUpdate) this.docSelectionOpUpdate = throttle(this._update, 1000).bind(this);
        const selection = this.context.selection;
        const textselection = this.context.selection.textSelection;
        this.docSelectionOpUpdate?.({
            select_page_id: selection.selectedPage?.id ?? "",
            select_shape_id_list: selection.selectedShapes.map((shape) => shape.id),
            hover_shape_id: selection.hoveredShape?.id,
            cursor_start: textselection.cursorStart,
            cursor_end: textselection.cursorEnd,
            cursor_at_before: textselection.cursorAtBefore,
        }).catch(err => { });
    }

    public addOnMessage(onMessage: (data: DocSelectionOpData) => void) {
        this.onMessageList.push(onMessage);
    }

    public removeOnMessage(onMessage: (data: DocSelectionOpData) => void) {
        const index = this.onMessageList.indexOf(onMessage);
        if (index >= 0) this.onMessageList.splice(index, 1);
    }
}