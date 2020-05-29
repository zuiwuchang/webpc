import { Component, OnInit, Inject } from '@angular/core';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
// CmdYes 確認操作
const CmdYes = 6
// CmdNo 取消操作
const CmdNo = 7
// CmdYesAll 覆蓋全部 重複檔案
const CmdYesAll = 9
// CmdSkip 跳過 重複檔案
const CmdSkip = 10
// CmdSkipAll 跳過全部 重複檔案
const CmdSkipAll = 11
@Component({
  selector: 'app-exist-choice',
  templateUrl: './exist-choice.component.html',
  styleUrls: ['./exist-choice.component.scss']
})
export class ExistChoiceComponent implements OnInit {
  constructor(
    private matDialogRef: MatDialogRef<ExistChoiceComponent>,
    @Inject(MAT_DIALOG_DATA) public filename: string, ) { }

  ngOnInit(): void {
  }
  onYes() {
    this.matDialogRef.close(CmdYes)
  }
  onYesAll() {
    this.matDialogRef.close(CmdYesAll)
  }
  onNo() {
    this.matDialogRef.close(CmdNo)
  }
  onSkip() {
    this.matDialogRef.close(CmdSkip)
  }
  onSkipAll() {
    this.matDialogRef.close(CmdSkipAll)
  }

}
