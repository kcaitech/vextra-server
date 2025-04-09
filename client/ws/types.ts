export enum DataTypes {
    Op = "op",
    Comment = "comment",
    Resource = "resource",
    Selection = "selection",
    DocUpload = "docupload",
    Bind = "bind",
    Start = "start",
    Heartbeat = "heartbeat"
}

export interface TransData {
    type: DataTypes,
    data_id: string,
    data: string,
    err?: string
}

export enum DocCommentOpType {
    Add = 0,
    Del,
    Update,
}

export interface Comment {
    id: string;
    content: string;
    user_id: string;
    create_time: number;
    update_time: number;
    page_id: string;
    shape_id?: string;
    position?: {
        x: number;
        y: number;
    };
}

export interface DocCommentOpData {
    type: DocCommentOpType;
    comment: Comment;
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
    // avatar?: string,
    // nickname?: string,
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
    // preVersion: number;
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

export interface IContext {
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
    lastRemoteCmdVersion: () => string | undefined;
}

export enum SelectionEvents {
    page_change = "page_change",
    shape_change = "shape_change",
    shape_hover_change = "shape_hover_change",
    text_change = "text_change"
}