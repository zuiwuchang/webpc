import { Component, OnInit, Inject } from '@angular/core';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';

@Component({
  selector: 'app-exist',
  templateUrl: './exist.component.html',
  styleUrls: ['./exist.component.scss']
})
export class ExistComponent implements OnInit {

  constructor(
    private matDialogRef: MatDialogRef<ExistComponent>,
    @Inject(MAT_DIALOG_DATA) public filename: string, ) { }

  ngOnInit(): void {
  }
  onSure() {
    this.matDialogRef.close(true)
  }
  onClose() {
    this.matDialogRef.close()
  }
}
