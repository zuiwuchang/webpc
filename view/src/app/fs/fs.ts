import { ServerAPI } from '../core/core/api';

const audios = new Set();
([
    '.mp3',
    '.aac',
    '.flac',
    '.ape',
    '.wav',
    '.ogg',
]).forEach(function (str: string) {
    audios.add(str)
});
const videos = new Set();
([
    '.webm',
    '.mp4', '.m4v',
    '.mov',
    '.avi',
    '.flv',
    '.wmv', '.asf',
    '.mpeg', '.mpg', '.vob',
    '.mkv',
    '.rm', '.rmvb',
]).forEach(function (str: string) {
    videos.add(str)
});
const images = new Set();
([
    '.gif',
    '.jpeg', '.jpg',
    '.bmp',
    '.png',
    '.svg', '.ico',
    '.webp',

]).forEach(function (str: string) {
    images.add(str)
});
const texts = new Set();
([
    '.txt', '.text',
    '.json', '.xml', '.yaml', '.ini',
    '.sh',
    '.bat', '.cmd', '.vbs',
    '.go', '.dart', '.py', '.py3', '.js', '.ts', '.c', '.cc', '.h', '.hpp', '.cpp',
]).forEach(function (str: string) {
    texts.add(str)
});
export enum FileType {
    Dir,
    Video,
    Audio,
    Image,
    Text,
    Binary,
}
export interface Dir {
    root: string
    read: boolean
    write: boolean
    shared: boolean
    dir: string
}
export class FileInfo {
    name: string
    mode: number
    size: number
    isDir: boolean

    filename: string
    root: string
    checked = false
    private _filetype = FileType.Binary

    constructor(root: string, dir: string, other: FileInfo) {
        this.name = other.name
        this.mode = other.mode
        this.size = other.size
        this.isDir = other.isDir
        if (dir.endsWith('/')) {
            this.filename = dir + other.name
        } else {
            this.filename = dir + '/' + other.name
        }
        this.root = root

        if (this.isDir) {
            this._filetype = FileType.Dir
        } else {
            const ext = this.ext.toLowerCase()
            if (videos.has(ext)) {
                this._filetype = FileType.Video
            } else if (audios.has(ext)) {
                this._filetype = FileType.Audio
            } else if (images.has(ext)) {
                this._filetype = FileType.Image
            } else if (texts.has(ext)) {
                this._filetype = FileType.Text
            } else {
                this._filetype = FileType.Binary
            }
        }
    }
    static compare(l: FileInfo, r: FileInfo): number {
        let lv = 0
        if (l.isDir) {
            lv = 1
        }
        let rv = 0
        if (r.isDir) {
            rv = 1
        }
        if (lv == rv) {
            if (l.name == r.name) {
                return 0
            }
            return l.name < r.name ? -1 : 1
        }
        return rv - lv
    }
    get ext(): string {
        const index = this.name.lastIndexOf('.')
        if (index == -1) {
            return ''
        }
        return this.name.substring(index)
    }
    get filetype(): FileType {
        return this._filetype
    }
    get isSupportUncompress(): boolean {
        if (this.isDir) {
            return false
        }
        const name = this.name.toLowerCase()
        return name.endsWith(`.tar.gz`) || name.endsWith(`.tar.bz2`)
    }
    get url(): string {
        switch (this._filetype) {
            case FileType.Dir:
                return '/fs/list'
            case FileType.Video:
                return '/fs/view/video'
            case FileType.Audio:
                return '/fs/view/audio'
            case FileType.Image:
                return '/fs/view/image'
            case FileType.Text:
                return '/fs/view/text'
        }
        return ''
    }
    get downloadURL(): string {
        if (this.isDir) {
            return ''
        }
        return ServerAPI.v1.fs.oneURL([this.root, this.filename])
    }
}