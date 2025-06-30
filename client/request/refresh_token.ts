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

function isTokenExpired(httpMgr: HttpMgr): boolean {
    const expRemain = httpMgr.token_remain;
    if (expRemain <= 60 * 1000) return true; // 前后端时间可能有差异，剩余1分钟也刷新
    return false;
}

export async function checkRefreshToken(httpMgr: HttpMgr) {
    if (httpMgr.token && isTokenExpired(httpMgr)) {
        const res = await refreshToken(httpMgr);
        if (res.code === HttpCode.StatusOK && res.data.token) {
            httpMgr.token = res.data.token;
        }
    }
}