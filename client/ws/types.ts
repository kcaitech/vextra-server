

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

export type DocCommentOpData = {
    type: DocCommentOpType
    comment: any
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
    avatar?: string,
    nickname?: string,
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
