
import { DataTypes, NetworkStatusType, TransData } from "./types";


interface LocalData {
    type: DataTypes,
    data_id: string,
    data: string | ArrayBuffer,
    retryCount: number
    resolve?: (value: { data?: any, buffer?: ArrayBuffer, err?: string }) => void
    // reject?: () => void
    timer?: ReturnType<typeof setTimeout>
}

function encodeBinaryData(str: string, buffer: ArrayBufferLike): ArrayBuffer {
    // 创建一个 TextEncoder 对象来编码字符串
    const encoder = new TextEncoder();
    // 编码字符串到 Uint8Array
    const uint8Str = encoder.encode(str);

    // 字符串的长度，以字节为单位
    const strLength = uint8Str.length;

    // 创建一个视图来存储字符串长度
    const lengthView = new DataView(new ArrayBuffer(4)); // 假设使用4字节表示长度
    lengthView.setUint32(0, strLength, true); // 设置长度，true表示小端序

    // 获取现有的 ArrayBuffer 的 Uint8Array 视图
    const uint8Buffer = new Uint8Array(buffer);

    // 计算总长度
    const totalLength = 4 + uint8Str.byteLength + uint8Buffer.byteLength;

    // 创建一个新的 Uint8Array
    const result = new Uint8Array(totalLength);

    // 使用 set 方法将长度信息的 Uint8Array 复制到结果数组中
    result.set(new Uint8Array(lengthView.buffer));
    // 使用 set 方法将字符串的 Uint8Array 复制到结果数组中
    result.set(uint8Str, 4);
    // 使用 set 方法将 buffer 的 Uint8Array 复制到结果数组中
    result.set(uint8Buffer, 4 + uint8Str.byteLength); // 从字符串结束的位置开始复制

    // 返回新的 ArrayBuffer
    return result.buffer;
}

function decodeBinaryData(receivedBuffer: ArrayBuffer) {
    const dataView = new DataView(receivedBuffer);

    // 读取前4个字节作为长度信息
    const strLength = dataView.getUint32(0, true); // 使用小端序

    // 提取字符串部分
    const strData = new Uint8Array(receivedBuffer, 4, strLength);
    const decoder = new TextDecoder('utf-8');
    const str = decoder.decode(strData);

    // 提取 ArrayBuffer 部分
    const bufferStart = 4 + strLength;
    const bufferData = new Uint8Array(receivedBuffer, bufferStart);
    const buffer = bufferData.buffer;

    return { str, buffer }
}

export class Connect {
    private ws?: WebSocket;
    // private binaryHandler: ((data: ArrayBuffer) => boolean)[] = []
    private dataHandler: { [key: string]: (data: any, binary?: ArrayBuffer) => void } = {}
    // private pendingDatas: PromisData[] = []
    private autoReconnect: boolean = true;
    private reconnectCount: number = 0;
    private connectTimer?: ReturnType<typeof setTimeout>;

    private data_id: number = 0;
    private promises: Map<string, LocalData> = new Map();


    private onChangeList: ((networkStatus: NetworkStatusType) => void)[] = []
    private _waitReady: ((val: boolean) => void)[] = []
    private _heartbeatInterval?: ReturnType<typeof setInterval>;

    private _wsUrl: string
    private _token?: string
    constructor(wsUrl: string, token?: string) {
        this._wsUrl = wsUrl
        this._token = token
    }

    private get localStorage() {
        if (typeof window !== 'undefined' && window.localStorage) {
            return window.localStorage
        }
        return undefined
    }

    private get token() {
        if (this._token) return this._token
        return this.localStorage?.getItem('token') ?? undefined
    }
    private set token(value: string | undefined) {
        this._token = value
        if (value) {
            this.localStorage?.setItem('token', value)
        } else {
            this.localStorage?.removeItem('token')
        }
    }

    setMessageHandler(type: string, handler: ((data: any, binary?: ArrayBuffer) => void) | undefined) {
        if (handler) this.dataHandler[type] = handler;
        else delete this.dataHandler[type]
    }

    // onBinary(handler: (data: ArrayBuffer) => boolean) {
    //     this.binaryHandler.push(handler)
    // }

    start(delay: number = 0) {
        if (this.ws) return;
        if (this.connectTimer) {
            clearTimeout(this.connectTimer)
            this.connectTimer = undefined
        }

        const connect = () => {
            console.log("connect")
            const tokenUrl = this._wsUrl + '?token=' + encodeURIComponent(this.token ?? "")
            const ws = new WebSocket(tokenUrl);
            ws.onclose = this.receiveClose.bind(this)
            ws.onerror = this.receiveError.bind(this);
            ws.onopen = this.receiveOpen.bind(this);
            ws.onmessage = this.receiveMessage.bind(this);
            ws.binaryType = 'arraybuffer'
            this.ws = ws;
        }

        if (delay === 0) {
            connect();
        } else {
            this.connectTimer = setTimeout(() => {
                this.connectTimer = undefined;
                connect()
            }, delay)
        }
    }

    private async asyncSend(data: LocalData, timeout: number): Promise<{ data?: any, buffer?: ArrayBuffer, err?: string }> {

        const promise = new Promise<{ data?: any, buffer?: ArrayBuffer, err?: string }>((resolve, reject) => {
            // data.reject = reject
            data.resolve = resolve
        })

        this.promises.set(data.data_id, data);
        const sendpack = () => {
            // console.log("ws send ", typeof data.data)
            this.ws?.send(data.data)
        }

        if (timeout) {
            const timeoutf = () => {
                // 超时
                data.timer = undefined
                if (data.retryCount <= 0) {
                    data.resolve?.({ err: "time out" });
                } else {
                    sendpack();
                    --data.retryCount
                    data.timer = setTimeout(timeoutf, timeout)
                }
            }
            data.timer = setTimeout(timeoutf, timeout)
        }

        sendpack();

        return promise
    }

    async send(type: DataTypes, data: Object, timeout: number = 500, retryCount: number = 3) {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return { err: 'ws not ready' };
        const data_id = "c" + (++this.data_id)
        const pack: LocalData = {
            type,
            data: JSON.stringify({
                type,
                data: JSON.stringify(data), // 需要单独序列化
                data_id
            }),
            data_id,
            retryCount,
        }

        return this.asyncSend(pack, timeout)
    }
    async sendBinary(type: DataTypes, header: Object, data: ArrayBufferLike, timeout: number, retryCount: number) {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return { err: 'ws not ready' };

        const data_id = "c" + (++this.data_id)
        const d = encodeBinaryData(JSON.stringify({
            type,
            data: JSON.stringify(header), // 需要单独序列化
            data_id,
        }), data)

        const pack: LocalData = {
            type,
            data: d,
            data_id,
            retryCount,
        }
        return this.asyncSend(pack, timeout)
    }

    private receiveMessage(ev: MessageEvent) {
        let json;
        let _buffer;
        if (typeof ev.data === 'string') {
            try {
                json = JSON.parse(ev.data) as TransData;
            } catch (error) {
                console.error('Failed to parse JSON:', error);
            }
        } else if (ev.data instanceof ArrayBuffer) {
            const { str, buffer } = decodeBinaryData(ev.data)
            _buffer = buffer;
            try {
                json = JSON.parse(str) as TransData;
            } catch (error) {
                console.error('Failed to parse JSON 1:', error);
            }
        } else {
            console.log('Received Unknown Type:', ev.data);
        }
        if (!json) return;

        const json_data = json.data && JSON.parse(json.data)
        if (json.type === DataTypes.Heartbeat) {
            return; // 无需处理
        }
        if (json.data_id.startsWith("c")) {
            const promise = this.promises.get(json.data_id)
            const err = json.err;
            if (err) console.log(err)
            if (promise) {
                this.promises.delete(json.data_id)
                if (promise.timer) clearTimeout(promise.timer)
                promise.resolve?.({ data: json_data, buffer: _buffer, err: json.err });
            }
        } else {
            const h = this.dataHandler[json.type];
            if (h) h(json_data, _buffer);
        }
    }

    // heartbeat
    private _heartbeat() {
        this.ws?.send(JSON.stringify({
            type: DataTypes.Heartbeat,
            data_id: ""
        }))
    }

    private receiveOpen(ev: Event) {

        if (!this._heartbeatInterval) {
            this._heartbeatInterval = setInterval(this._heartbeat.bind(this), 30 * 1000); // 30s
        }

        console.log("ws receive open")
        // this.pendingDatas.forEach(d => {
        //     d.send(); // todo
        // })
        // this.pendingDatas = []
        this.reconnectCount = 0
        this._waitReady.forEach(l => l(true))
        this.onChangeList.forEach(l => l(NetworkStatusType.Online))
    }

    private receiveError(ev: Event) {
        console.log("ws receive error")
        console.error(ev)
    }

    private receiveClose(ev: CloseEvent) {
        console.log("ws receive close")
        this.ws = undefined
        this._waitReady.forEach(l => l(false))
        this.onChangeList.forEach(l => l(NetworkStatusType.Offline))
        if (this.autoReconnect) {
            ++this.reconnectCount;
            this.start(1000 * this.reconnectCount)
        }
    }

    close() {
        if (this._heartbeatInterval) {
            clearInterval(this._heartbeatInterval)
            this._heartbeatInterval = undefined
        }
        if (this.connectTimer) {
            clearTimeout(this.connectTimer)
            this.connectTimer = undefined
        }
        console.log("close connect")
        this.autoReconnect = false;
        const ws = this.ws;
        this.ws = undefined;
        if (ws) ws.close();
    }

    get isReady(): boolean {
        return this.ws && this.ws.readyState === WebSocket.OPEN || false;
    }

    public addOnChange(onChange: (networkStatus: NetworkStatusType) => void) {
        this.onChangeList.push(onChange)
    }

    public removeOnChange(onChange: (networkStatus: NetworkStatusType) => void) {
        const index = this.onChangeList.indexOf(onChange)
        if (index >= 0) this.onChangeList.splice(index, 1);
    }

    async waitReady() {
        if (this.isReady) return true;
        return new Promise<boolean>((resolve: (val: boolean) => void) => {
            this._waitReady.push(resolve)
        })
    }
}


export abstract class ConnectClient {
    private connect: Connect;
    private type: DataTypes;
    constructor(connect: Connect, type: DataTypes) {
        this.connect = connect;
        this.type = type;
        connect.setMessageHandler(type, this.onMessage.bind(this))
        connect.start();
    }
    get isConnected() {
        return this.connect.isReady
    }

    async waitReady() { // todo 长时间不用关闭连接，再次wait时再connect或者一段时间后再连接上去收协作消息
        return this.connect.waitReady()
    }

    hasConnected(): boolean {
        return this.connect.isReady
    }

    async send(data: any, timeout = 1000, retryCount = 3) {
        return this.connect.send(this.type, data, timeout, retryCount)
    }

    async sendBinary(header: any, data: ArrayBufferLike, timeout = 1000, retryCount = 3) {
        return this.connect.sendBinary(this.type, header, data, timeout, retryCount)
    }

    abstract onMessage(data: any): void;
}