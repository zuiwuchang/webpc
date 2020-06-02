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
  const vals = new Array<number>()
  const hash = new SparkMD5()
  fileReader.onload = function (e) {
    try {
      const spark = new SparkMD5.ArrayBuffer()
      spark.append(e.target.result);
      const val = spark.end()
      vals.push(val)

      hash.append(val)

      index++
      if (index < requests.length) {
        loadNext()
      } else {
        postMessage({
          vals: vals,
          hash: hash.end(),
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
