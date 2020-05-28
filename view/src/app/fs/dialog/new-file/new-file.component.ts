import { Component, OnInit, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ServerAPI } from 'src/app/core/core/api';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { FileInfo, Dir } from '../../fs';

@Component({
  selector: 'app-new-file',
  templateUrl: './new-file.component.html',
  styleUrls: ['./new-file.component.scss']
})
export class NewFileComponent implements OnInit {
  constructor(private httpClient: HttpClient,
    private toasterService: ToasterService,
    private i18nService: I18nService,
    private matDialogRef: MatDialogRef<NewFileComponent>,
    @Inject(MAT_DIALOG_DATA) public data: Dir,
  ) {
  }
  private _disabled: boolean
  get disabled(): boolean {
    return this._disabled
  }
  ngOnInit(): void {
    this.name = this.i18nService.get('New File')
  }
  name: string
  onSubmit() {
    this._disabled = true
    ServerAPI.v1.fs.postOne<FileInfo>(this.httpClient, [this.data.root, this.data.dir], {
      name: this.name,
    }).then((data) => {
      this.toasterService.pop('success', undefined, this.i18nService.get(`New File Success`))
      const node = new FileInfo(this.data.root, this.data.dir, data)
      this.matDialogRef.close(node)
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
