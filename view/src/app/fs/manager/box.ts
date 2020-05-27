export class Point {
    constructor(public x: number, public y: number) {
    }
    toView(): Point {
        if (document.compatMode == "BackCompat") {
            this.x -= document.body.scrollLeft
            this.y -= document.body.scrollTop
        } else {
            this.x -= document.documentElement.scrollLeft
            this.y -= document.documentElement.scrollTop
        }
        return this
    }
}
// 有效範圍
export class Box {
    private _p0: Point
    private _p1: Point
    start: Point
    stop: Point
    setRange(element) {
        this._p0 = getViewPoint(element)
        this._p1 = new Point(this._p0.x + element.offsetWidth, this._p0.y + element.offsetHeight)
    }
    private _fixStart() {
        if (this.start.x < this._p0.x) {
            this.start.x = this._p0.x
        } else if (this.start.x > this._p1.x) {
            this.start.x = this._p1.x
        }

        if (this.start.y < this._p0.y) {
            this.start.y = this._p0.y
        } else if (this.start.y > this._p1.y) {
            this.start.y = this._p1.y
        }
    }

    private _fixStop() {
        if (this.stop.x < this._p0.x) {
            this.stop.x = this._p0.x
        } else if (this.stop.x > this._p1.x) {
            this.stop.x = this._p1.x
        }

        if (this.stop.y < this._p0.y) {
            this.stop.y = this._p0.y
        } else if (this.stop.y > this._p1.y) {
            this.stop.y = this._p1.y
        }
    }
    calculate() {
        if (!this.start || !this.stop) {
            return
        }
        if (this._p0 && this._p1) {
            this._fixStart()
            this._fixStop()
        }
        this.x = Math.min(this.start.x, this.stop.x)
        this.y = Math.min(this.start.y, this.stop.y)
        this.w = Math.abs(this.start.x - this.stop.x)
        this.h = Math.abs(this.start.y - this.stop.y)
    }
    x = 0
    y = 0
    w = 0
    h = 0
    reset() {
        this.x = 0
        this.y = 0
        this.w = 0
        this.h = 0
        this._p0 = null
        this._p1 = null
        this.start = null
        this.stop = null
    }

    checked(doc: Document): Array<number> {
        const result = new Array<number>()
        const nodes = doc.childNodes

        if (nodes && nodes.length > 0) {
            let parent: any
            for (let i = 0; i < nodes.length; i++) {
                let node = (nodes[i] as any)
                if (!node || !node.querySelector) {
                    continue
                }
                node = node.querySelector('.wrapper')
                if (!node) {
                    continue
                }
                const l = getViewPoint(node)
                const r = new Point(l.x + node.offsetWidth, l.y + node.offsetHeight)
                const ok = this.testView(l, r)
                if (ok) {
                    result.push(i)
                }
            }
        }
        return result
    }
    testView(l: Point, r: Point): boolean {
        if (r.x < this.x || l.x > (this.x + this.w)) {
            return false
        }
        if (r.y < this.y || l.y > (this.y + this.h)) {
            return false
        }
        return true
    }
}
export function getPagePoint(element): Point {
    let x = 0
    let y = 0
    while (element) {
        x += element.offsetLeft + element.clientLeft
        y += element.offsetTop + element.clientTop
        element = element.offsetParent
    }
    return new Point(x, y)
}

export function getViewPoint(element): Point {
    return getPagePoint(element).toView()
}