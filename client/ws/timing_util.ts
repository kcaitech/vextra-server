/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

type UnwrappedPromise<T> = T extends Promise<infer U> ? U : T

// 节流
export function throttle<T extends (...args: any[]) => any>(func: T, delay: number): (...funcArgs: Parameters<T>) => Promise<UnwrappedPromise<ReturnType<T>>> {
    let timer: ReturnType<typeof setTimeout> | undefined = undefined
    let cancel: (reason?: any) => void = () => {}
    let previousRunTime: number = 0
    return function (this: any, ...args: Parameters<T>) {
        if (timer) {
            clearTimeout(timer)
            timer = undefined
            cancel("cancel")
        }
        const now = Date.now()
        const remaining = delay - (now - previousRunTime)
        if (remaining <= 0) {
            previousRunTime = now
            return func.apply(this, args) // eslint-disable-line prefer-spread
        }
        return new Promise((resolve, reject) => {
            timer = setTimeout(() => {
                timer = undefined
                previousRunTime = Date.now()
                resolve(func.apply(this, args)) // eslint-disable-line prefer-spread
            }, remaining) as any
            cancel = reject
        })
    }
}

// 防抖
export function debounce<T extends (...args: any[]) => any>(func: T, delay: number): (...funcArgs: Parameters<T>) => Promise<UnwrappedPromise<ReturnType<T>>> {
    let timer: ReturnType<typeof setTimeout> | undefined = undefined
    let cancel: (reason?: any) => void = () => {}
    return function (this: any, ...args: Parameters<T>) {
        if (timer) {
            clearTimeout(timer)
            timer = undefined
            cancel("cancel")
        }
        return new Promise((resolve, reject) => {
            timer = setTimeout(() => {
                timer = undefined
                resolve(func.apply(this, args)) // eslint-disable-line prefer-spread
            }, delay) as any
            cancel = reject
        })
    }
}
