import { Component, OnInit, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ServerAPI } from 'src/app/core/core/api';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { FileInfo, Dir } from '../../fs';
interface Target {
  dir: Dir
  source: Array<FileInfo>
}
@Component({
  selector: 'app-remove',
  templateUrl: './remove.component.html',
  styleUrls: ['./remove.component.scss']
})
export class RemoveComponent implements OnInit {

  constructor(private httpClient: HttpClient,
    private toasterService: ToasterService,
    private i18nService: I18nService,
    private matDialogRef: MatDialogRef<RemoveComponent>,
    @Inject(MAT_DIALOG_DATA) public target: Target,
  ) { }

  ngOnInit(): void {
  }
  private _disabled: boolean
  get disabled(): boolean {
    return this._disabled
  }
  onSubmit() {
    this._disabled = true
    ServerAPI.v1.fs.deleteOne(this.httpClient, [this.target.dir.root, this.target.dir.dir], {
      params: {
        names: this.target.source.map<string>((node) => node.name),
      },
    }).then(() => {
      this.toasterService.pop('success', undefined, this.i18nService.get(`File deleted`))
      this.matDialogRef.close(true)
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
