/// <reference lib="webworker" />

// import { buf } from "crc-32";
import * as SparkMD5 from 'spark-md5';
interface Data {
  file: File
  start: number
  end: number
}
addEventListener('message', ({ data }) => {
  const requests: Array<Data> = data
  const fileReader = new FileReader()
  let index = 0
  const vals = new Array<string>()
  fileReader.onload = function (e) {
    try {
      // vals.push(buf(new Uint8Array(e.target.result as ArrayBuffer)))
      const spark = new SparkMD5.ArrayBuffer()
      spark.append(e.target.result);
      vals.push(spark.end())

      index++
      if (index < requests.length) {
        loadNext()
      } else {
        postMessage({
          vals: vals,
        })
      }
    } catch (e) {
      console.warn(e)
      postMessage({
        error: e,
      })
    }
  }
  fileReader.onerror = function (evt) {
    console.warn(evt)
    postMessage({
      error: 'FileReader error',
    })
  }
  function loadNext() {
    const request = requests[index]
    fileReader.readAsArrayBuffer(request.file.slice(request.start, request.end))
  }
  loadNext()
});
