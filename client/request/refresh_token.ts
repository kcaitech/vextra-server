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

// 用于跟踪 refresh token 请求的 Promise
let refreshTokenPromise: Promise<any> | null = null;
export async function checkRefreshToken(httpMgr: HttpMgr) {

    if (!httpMgr.token || !isTokenExpired(httpMgr)) {
        return;
    }

    // 如果已经有 refresh token 请求在进行中，等待它完成
    if (refreshTokenPromise) {
        return await refreshTokenPromise;
    }
    
    // 创建新的 refresh token 请求
    refreshTokenPromise = (async () => {
        try {
            const res = await refreshToken(httpMgr);
            if (res.code === HttpCode.StatusOK && res.data.token) {
                httpMgr.token = res.data.token;
            }
            return res;
        } catch (err) {
            console.error('RefreshToken请求失败:', err);
            // 刷新失败时清除 token
            httpMgr.token = undefined;
            throw err;
        } finally {
            // 无论成功还是失败，都要重置 Promise
            refreshTokenPromise = null;
        }
    })();
    
    return await refreshTokenPromise;
}