/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

import { z } from "zod";
import { CommentItem } from "../request/comment";
import { AccessKeyInfoSchema, DocumentInfoSchemaEx, UserInfo } from "../common/types";

export enum DataTypes {
    Op = "op",
    Comment = "comment",
    Resource = "resource",
    Thumbnail = "thumbnail",
    Selection = "selection",
    DocUpload = "docupload",
    Bind = "bind",
    Start = "start",
    Heartbeat = "heartbeat",
    GenerateVersion = "generateVersion",
}

export interface TransData {
    type: DataTypes;
    data_id: string;
    data: string;
    msg: string;
    code: number;
}

export enum DocCommentOpType {
    Add = 0,
    Del,
    Update,
}

export interface DocCommentOpData {
    type: DocCommentOpType;
    comment: CommentItem;
    user?: UserInfo;
    create_at?: string;
}

export type ResourceHeader = {
    name: string,
}

export enum DocSelectionOpType {
    Update = 0,
    Exit,
}

export type DocSelectionData = {
    select_page_id: string,
    select_shape_id_list: string[],
    hover_shape_id?: string,
    cursor_start?: number,
    cursor_end?: number,
    cursor_at_before?: boolean,
    // 以下字段仅读取时有效
    user_id?: string,
    permission?: number,
    user?: UserInfo,
    enter_time?: number,
}

export type DocSelectionOpData = {
    type: DocSelectionOpType,
    user_id: string,
    data: DocSelectionData,
}

export enum NetworkStatusType {
    Online = 0,
    Offline,
}

export declare enum OpType {
    None = 0,
    Array = 1,
    Idset = 2,
    CrdtArr = 3,
    CrdtTree = 4
}
export interface Op {
    id: string;
    path: string[];
    type: OpType;
}

export interface Cmd {
    id: string;
    baseVer: number;
    batchId: string;
    ops: Op[];
    isRecovery: boolean;
    description: string;
    time: number;
    posttime: number;
    dataFmtVer: string;
    version: number;
}
export interface OpItem {
    op: Op;
    cmd: Cmd;
}

export interface ResourceData {
    header: ResourceHeader;
    data: ArrayBuffer;
}

export interface DocUploadData {
    file_id: string;
    file_name: string;
    file_size: number;
    file_type: string;
    upload_time: number;
    user_id: string;
}

export interface BindData {
    user_id: string;
    doc_id: string;
    permission: number;
}

export interface StartData {
    doc_id: string;
    version: string;
}

export interface HeartbeatData {
    timestamp: number;
}

export interface IClientContext {
    selection: {
        selectedPage?: { id: string };
        selectedShapes: Array<{ id: string }>;
        hoveredShape?: { id: string };
        textSelection: {
            cursorStart?: number;
            cursorEnd?: number;
            cursorAtBefore?: boolean;
        };
        watch: (callback: (type: string) => void) => void;
    };
    lastRemoteCmdVersion: () => number | undefined;
}

export enum SelectionEvents {
    page_change = "page_change",
    shape_change = "shape_change",
    shape_hover_change = "shape_hover_change",
    text_change = "text_change"
}

export type VersionData = {
    document_id: string;
    version_id: string;
    version_start_with: number;
}


export const BaseResponseSchema = z.object({
    // type: z.string(),
	// data_id: z.string(),
	// // data: z.string().optional(), // 需要解析成具体的类型
	// msg: z.string().optional(),
	code: z.number().optional(),
})


export const BindRespDataSchema = z.object({
    doc_info: DocumentInfoSchemaEx,
    access_key: AccessKeyInfoSchema,
})

export type BindResponseData = z.infer<typeof BindRespDataSchema>;

export const BindResponseSchema = BaseResponseSchema.extend({
    data: BindRespDataSchema,
})

export type BindResponse = z.infer<typeof BindResponseSchema>;