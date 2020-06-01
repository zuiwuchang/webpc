/// <reference lib="webworker" />

import { buf } from "crc-32";
interface Data {
  file: File
  start: number
  end: number
}
addEventListener('message', ({ data }) => {
  const request: Data = data
  request.file.slice(request.start, request.end).arrayBuffer().then((data) => {
    try {
      const hash = buf(new Uint8Array(data), 0)
      postMessage({
        val: hash,
      })
    } catch (e) {
      postMessage({
        error: e,
      })
    }
  }).catch((e) => {
    postMessage({
      error: e,
    })
  })
});
