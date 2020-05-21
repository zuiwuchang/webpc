import { RESTful } from './restful';
const root = '/api'

export const ServerAPI = {
    v1: {
        session: new RESTful(`${root}/v1/session`),
        version: new RESTful(`${root}/v1/version`),
        users: new RESTful(`${root}/v1/users`),
    },
}
export function getWebSocketAddr(path: string): string {
    const location = document.location
    let addr: string
    if (location.protocol == "https") {
        addr = `wss://${location.hostname}`
        if (location.port == "") {
            addr += ":443"
        } else {
            addr += `:${location.port}`
        }
    } else {
        addr = `ws://${location.hostname}`
        if (location.port == "") {
            addr += ":80"
        } else {
            addr += `:${location.port}`
        }
    }
    return `${addr}${path}`
}