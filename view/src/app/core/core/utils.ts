export const KB = 1024
export const MB = 1024 * KB
export const GB = 1024 * MB
export const TB = 1024 * GB
export function sizeString(val: number): string {
    if (typeof val !== "number" || isNaN(val)) {
        return '0b'
    }
    val = Math.floor(val)
    if (val < 1) {
        return '0b'
    }
    const strs = new Array<string>()
    if (val > TB) {
        const tmp = Math.floor(val / TB)
        strs.push(`${tmp}t`)
        val -= tmp * TB
    }
    if (val > GB) {
        const tmp = Math.floor(val / GB)
        strs.push(`${tmp}g`)
        val -= tmp * GB
    }
    if (val > MB) {
        const tmp = Math.floor(val / MB)
        strs.push(`${tmp}m`)
        val -= tmp * MB
    }
    if (val > KB) {
        const tmp = Math.floor(val / KB)
        strs.push(`${tmp}k`)
        val -= tmp * KB
    }
    if (val > 0) {
        strs.push(`${val}b`)
    }
    return strs.join('')
}