export class Shell {
    // shell id
    id: string
    // shell 顯示名稱
    name: string
    // 是否 附加 websocket
    attached: boolean

    static compare(l: Shell, r: Shell): number {
        if (l.name != r.name) {
            return l.name > r.name ? 1 : -1
        }
        if (l.id != r.id) {
            return l.id > r.id ? 1 : -1
        }
        return 0
    }
}
