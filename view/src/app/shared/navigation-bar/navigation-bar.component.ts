import { Component, OnInit, OnDestroy } from '@angular/core';
import { Session, SessionService } from 'src/app/core/session/session.service';
import { Subscription, Subject } from 'rxjs';
import { MatDialog } from '@angular/material/dialog';
import { LoginComponent } from '../login/login.component';
import { PasswordComponent } from '../password/password.component';
import { FullscreenService } from 'src/app/core/fullscreen/fullscreen.service';
import { takeUntil } from 'rxjs/operators';
@Component({
  selector: 'shared-navigation-bar',
  templateUrl: './navigation-bar.component.html',
  styleUrls: ['./navigation-bar.component.scss'],
})
export class NavigationBarComponent implements OnInit, OnDestroy {
  constructor(
    private sessionService: SessionService,
    private matDialog: MatDialog,
    private fullscreenService: FullscreenService,
  ) { }
  private _ready = false;
  get ready(): boolean {
    return this._ready;
  }
  private _session: Session;
  get session(): Session {
    return this._session
  }
  private _closeSubject = new Subject<boolean>()
  fullscreen = false
  ngOnInit(): void {
    this.sessionService.ready.then((data) => {
      this._ready = data
    })
    this.sessionService.observable.pipe(
      takeUntil(this._closeSubject),
    ).subscribe((data) => {
      this._session = data
    });
    this.fullscreenService.observable.pipe(
      takeUntil(this._closeSubject),
    ).subscribe((data) => {
      this.fullscreen = data
    })
  }
  ngOnDestroy(): void {
    this._closeSubject.next(true)
    this._closeSubject.complete()
  }
  onClickPassword() {
    this.matDialog.open(PasswordComponent, {
      disableClose: true,
    })
  }
  onClickLogin() {
    this.matDialog.open(LoginComponent)
  }
  onClickLogout() {
    this.sessionService.logout();
  }
}
