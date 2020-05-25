import { Component, OnInit, Input } from '@angular/core';
import { FileInfo, FileType } from '../fs';
@Component({
  selector: 'fs-file',
  templateUrl: './file.component.html',
  styleUrls: ['./file.component.scss']
})
export class FileComponent implements OnInit {

  constructor() { }
  @Input()
  source: FileInfo
  ngOnInit(): void {
  }
  get icon(): string {
    if (this.source) {
      switch (this.source.filetype) {
        case FileType.Dir:
          return 'folder'
        case FileType.Video:
          return 'movie_creation'
        case FileType.Audio:
          return 'audiotrack'
        case FileType.Image:
          return 'insert_photo'
        case FileType.Text:
          return 'event_note'
      }
    }
    return 'insert_drive_file'
  }
  onContextmenu(evt: Event) {
    evt.stopPropagation()
    console.log('menu', this.source)
    return false
  }
  onClick(evt: Event) {
    evt.stopPropagation()
    console.log('click', this.source)
    return false
  }
  onDbclick() {
    console.log('dbclick', this.source)
  }
}
