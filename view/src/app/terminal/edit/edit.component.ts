import { Component, OnInit, OnDestroy, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ServerAPI } from 'src/app/core/core/api';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { Shell } from '../shell';
import { isString } from 'util';

@Component({
  selector: 'app-edit',
  templateUrl: './edit.component.html',
  styleUrls: ['./edit.component.scss']
})
export class EditComponent implements OnInit, OnDestroy {
  constructor(private httpClient: HttpClient,
    private toasterService: ToasterService,
    private i18nService: I18nService,
    private matDialogRef: MatDialogRef<EditComponent>,
    @Inject(MAT_DIALOG_DATA) public data: Shell,
  ) {
  }
  private _closed = false
  private _disabled: boolean
  get disabled(): boolean {
    return this._disabled
  }
  name: string
  ngOnInit(): void {
    this.name = this.data.name
  }
  ngOnDestroy() {
    this._closed = true
  }
  get isNotChanged(): boolean {
    let l = ''
    if (isString(this.name)) {
      l = this.name.trim()
    }
    let r = ''
    if (isString(this.data.name)) {
      r = this.data.name.trim()
    }
    return l == r
  }
  onSave() {
    ServerAPI.v1.shells.patchOne(this.httpClient, this.data.id, `name`, {
      name: this.name.trim(),
    }).then(() => {
      if (this._closed) {
        return
      }
      this.toasterService.pop('success',
        this.i18nService.get('success'),
        this.i18nService.get('change terminal completed'),
      )
      this.data.name = this.name
      this.matDialogRef.close()
    }, (e) => {
      if (this._closed) {
        return
      }
      console.warn(e)
      this.toasterService.pop('error',
        this.i18nService.get('error'),
        e,
      )
    }).finally(() => {
      this._disabled = false
    })
  }
  onClose() {
    this.matDialogRef.close()
  }
}
