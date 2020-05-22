import { Component, OnInit, OnDestroy, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ServerAPI } from 'src/app/core/core/api';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { sha512 } from 'js-sha512';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';

@Component({
  selector: 'app-password',
  templateUrl: './password.component.html',
  styleUrls: ['./password.component.scss']
})
export class PasswordComponent implements OnInit, OnDestroy {
  constructor(private httpClient: HttpClient,
    private toasterService: ToasterService,
    private i18nService: I18nService,
    private matDialogRef: MatDialogRef<PasswordComponent>,
  ) {
  }
  private _closed = false
  private _disabled: boolean
  get disabled(): boolean {
    return this._disabled
  }
  old: string
  val: string
  ngOnInit(): void {
  }
  ngOnDestroy() {
    this._closed = true
  }
  onSave() {
    this._disabled = true
    const old = sha512(this.old).toString()
    const val = sha512(this.val).toString()
    ServerAPI.v1.session.patch(this.httpClient, `password`, {
      old: old,
      val: val,
    }).then(() => {
      if (this._closed) {
        return
      }
      this.toasterService.pop('success',
        this.i18nService.get('success'),
        this.i18nService.get('change password completed'),
      )
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
