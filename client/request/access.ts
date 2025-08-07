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
import { HttpMgr } from "./http";
import { BaseResponseSchema } from "./types";

export const AccessListResponseItemSchema = z.object({
    key: z.string(),
    priority_mask: z.number(),
    resource_mask: z.number(),
    type: z.number(),
    resource_id: z.string(),
})

export const AccessListResponseSchema = BaseResponseSchema.extend({
    data: z.array(AccessListResponseItemSchema),
});

export type AccessListResponse = z.infer<typeof AccessListResponseSchema>;

export const AccessCreateResponseSchema = BaseResponseSchema.extend({
    data: z.object({
        access_key: z.string(),
        access_secret: z.string(),
    }),
});

export type AccessCreateResponse = z.infer<typeof AccessCreateResponseSchema>;

export const AccessUpdateResponseSchema = BaseResponseSchema.extend({
    data: z.object({
        message: z.string(),
    }),
});

export type AccessUpdateResponse = z.infer<typeof AccessUpdateResponseSchema>;

export const AccessDeleteResponseSchema = BaseResponseSchema.extend({
    data: z.object({
        message: z.string(),
    }),
});
export type AccessDeleteResponse = z.infer<typeof AccessDeleteResponseSchema>;

export const AccessTokenResponseSchema = BaseResponseSchema.extend({
    data: z.object({
        token: z.string(),
    }),
});
export type AccessTokenResponse = z.infer<typeof AccessTokenResponseSchema>;

export class AccessAPI {
    private http: HttpMgr

    constructor(http: HttpMgr) {
        this.http = http
    }

    async getAccessList(): Promise<AccessListResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/access/list',
            method: 'get',
        });
        return AccessListResponseSchema.parse(result);
    }

    async createAccessKey(params: {
        priority?: string[];
        document?: string[];
        project?: string[];
        team?: string[];
        user?: boolean;
    }): Promise<AccessCreateResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/access/create',
            method: 'post',
            data: params,
        });
        return AccessCreateResponseSchema.parse(result);
    }

    async updateAccessKey(params: {
        access_key: string;
        priority?: string[];
        document?: string[];
        project?: string[];
        team?: string[];
        user?: boolean;
    }): Promise<AccessUpdateResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/access/update',
            method: 'post',
            data: params,
        });
        return AccessUpdateResponseSchema.parse(result);
    }

    async deleteAccessKey(params: {
        access_key: string;
    }): Promise<AccessDeleteResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/access/delete',
            method: 'post',
            data: params,
        });
        return AccessDeleteResponseSchema.parse(result);
    }

    async getAccessToken(params: {
        access_key: string;
        access_secret: string;
    }): Promise<AccessTokenResponse> {
        await this.http.refresh_token();
        const result = await this.http.request({
            url: '/access/token',
            method: 'post',
            data: params,
        });
        return AccessTokenResponseSchema.parse(result);
    }
}