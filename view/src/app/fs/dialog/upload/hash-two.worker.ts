/// <reference lib="webworker" />

import { buf } from "crc-32";
// import * as md5 from "js-md5";
interface Data {
    file: File
    start: number
    end: number
}
addEventListener('message', async ({ data }) => {
    const requests: Array<Data> = data
    try {
        const val = new Array<number>()
        for (let i = 0; i < requests.length; i++) {
            const request = requests[i]
            const data = await request.file.slice(request.start, request.end).arrayBuffer()
            const hash = buf(new Uint8Array(data), 0)
            // const hash = md5(data)
            val.push(hash)
        }
        postMessage({
            val: val,
        })
    } catch (e) {
        postMessage({
            error: e,
        })
    }
});
