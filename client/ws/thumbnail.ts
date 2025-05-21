import { Connect, ConnectClient } from "./connect";
import { DataTypes } from "./types";

export class Thumbnail extends ConnectClient {
    constructor(connect: Connect) {
        super(connect, DataTypes.Thumbnail)
    }

    onMessage(data: any): void {
    }

    public async upload(name: string, contentType: string, data: ArrayBuffer): Promise<boolean> {
        await this.waitReady()
        return !!this.sendBinary({ name: name, contentType }, data, 1000, 3);
    }
}