import { Component, OnInit, Input, Output, EventEmitter } from '@angular/core';
import { FileInfo, FileType } from '../fs';
export interface NativeEvent extends Event {
  ctrlKey: boolean
  shiftKey: boolean
}
export interface CheckEvent {
  target: FileInfo
  event: NativeEvent
}
@Component({
  selector: 'fs-file',
  templateUrl: './file.component.html',
  styleUrls: ['./file.component.scss']
})
export class FileComponent implements OnInit {

  constructor() { }
  @Input()
  source: FileInfo
  @Output()
  checkChange = new EventEmitter<CheckEvent>()
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
  onContextmenu(evt: NativeEvent) {
    evt.stopPropagation()
    this.checkChange.emit({
      event: evt,
      target: this.source,
    })
    console.log('menu', this.source)
    return false
  }
  onClick(evt: NativeEvent) {
    evt.stopPropagation()
    this.checkChange.emit({
      event: evt,
      target: this.source,
    })
    return false
  }
  onDbclick() {
    console.log('dbclick', this.source)
  }
}
