

import * as base64 from "js-base64";
import { HttpMgr } from "./http";
import { BaseResponseSchema } from "./types";
import { z } from "zod";
import { HttpCode } from "./httpcode";

const RefreshTokenSchema = BaseResponseSchema.extend({
    data: z.object({
        token: z.string()
    })
})

export async function refreshToken(httpMgr: HttpMgr) {
    const result = await httpMgr.request({
        url: 'auth/refresh_token',
        method: 'post',
    })
    try {
        return RefreshTokenSchema.parse(result);
    } catch (error) {
        console.error('RefreshToken数据校验失败:', error);
        throw error;
    }
}

function getTokenExpireRemain(token: string): number {
    try {
        if (!token) return 0;
        const jwtSplitRes = token.split(".");
        if (jwtSplitRes.length !== 3) return 0; // jwt格式不正确
        const r = base64.decode(jwtSplitRes[1]);
        const jwtPayload = JSON.parse(r);
        const expRemain = (jwtPayload.exp ?? 0) * 1000 - Date.now();
        return expRemain;
    } catch (e) {
        console.log("parse jwt error", e);
        return 0;
    }
}

function isTokenExpired(token: string): boolean {
    const expRemain = getTokenExpireRemain(token);
    // console.log('isTokenExpired expRemain', expRemain / 1000 / 60, 'minutes')
    if (expRemain <= 60 * 1000) return true; // 前后端时间可能有差异，剩余1分钟也刷新
    return false;
}

export async function checkRefreshToken(httpMgr: HttpMgr) {
    const token = httpMgr.token;
    if (token && isTokenExpired(token)) {
        const res = await refreshToken(httpMgr);
        if (res.code === HttpCode.StatusOK && res.data.token) {
            httpMgr.token = res.data.token;
        }
    }
}