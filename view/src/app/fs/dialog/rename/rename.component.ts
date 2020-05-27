import { Component, OnInit, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ServerAPI } from 'src/app/core/core/api';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { FileInfo } from '../../fs';

@Component({
  selector: 'app-rename',
  templateUrl: './rename.component.html',
  styleUrls: ['./rename.component.scss']
})
export class RenameComponent implements OnInit {
  constructor(private httpClient: HttpClient,
    private toasterService: ToasterService,
    private i18nService: I18nService,
    private matDialogRef: MatDialogRef<RenameComponent>,
    @Inject(MAT_DIALOG_DATA) public data: FileInfo,
  ) {
  }
  private _disabled: boolean
  get disabled(): boolean {
    return this._disabled
  }
  ngOnInit(): void {
    this.name = this.data.name
  }
  name: string
  get isNotChanged(): boolean {
    return this.name == this.data.name
  }
  onSave() {
    this._disabled = true
    ServerAPI.v1.fs.patchOne<string>(this.httpClient, [this.data.root, this.data.filename], 'name', {
      val: this.name,
    }).then((name) => {
      this.name = name
    }, (e) => {
      this.toasterService.pop('error', undefined, e)
    }).finally(() => {
      this._disabled = false
    })
  }
  onClose() {
    this.matDialogRef.close()
  }
}
