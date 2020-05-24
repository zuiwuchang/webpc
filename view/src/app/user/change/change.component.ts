import { Component, OnInit, OnDestroy, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ServerAPI } from 'src/app/core/core/api';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { User } from '../user';

@Component({
  selector: 'app-change',
  templateUrl: './change.component.html',
  styleUrls: ['./change.component.scss']
})
export class ChangeComponent implements OnInit, OnDestroy {
  constructor(private httpClient: HttpClient,
    private toasterService: ToasterService,
    private i18nService: I18nService,
    private matDialogRef: MatDialogRef<ChangeComponent>,
    @Inject(MAT_DIALOG_DATA) public data: User,
  ) {
  }
  private _closed = false
  private _disabled: boolean
  get disabled(): boolean {
    return this._disabled
  }
  shell: boolean
  read: boolean
  write: boolean
  root: boolean
  ngOnInit(): void {
    if (this.data.shell) {
      this.shell = true
    }
    if (this.data.read) {
      this.read = true
    }
    if (this.data.write) {
      this.write = true
    }
    if (this.data.root) {
      this.root = true
    }
  }
  ngOnDestroy() {
    this._closed = true
  }
  get isNotChanged(): boolean {
    return !this.shell == !this.data.shell &&
      !this.read == !this.data.read &&
      !this.write == !this.data.write &&
      !this.root == !this.data.root
  }
  onSave() {
    this._disabled = true
    ServerAPI.v1.users.patchOne(this.httpClient, this.data.name, `change`, {
      shell: this.shell,
      read: this.read,
      write: this.write,
      root: this.root,
    }).then(() => {
      if (this._closed) {
        return
      }
      this.toasterService.pop('success',
        this.i18nService.get('success'),
        this.i18nService.get('change authorization completed'),
      )
      this.data.shell = this.shell
      this.data.read = this.read
      this.data.write = this.write
      this.data.root = this.root
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
