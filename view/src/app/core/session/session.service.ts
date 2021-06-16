import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { Mutex, Completer } from '../core/completer';
import { HttpClient } from '@angular/common/http';
import { ServerAPI } from '../core/api';
import { Exception } from '../core/exception';

export class Session {
  name: string
  shell: boolean
  read: boolean
  write: boolean
  root: boolean
}
@Injectable({
  providedIn: 'root'
})
export class SessionService {
  constructor(private httpClient: HttpClient) {
    // 恢復 session
    this._restore();
  }
  private readonly _mutex = new Mutex()
  private readonly _subject = new BehaviorSubject<Session>(null)
  private readonly _ready = new Completer<boolean>()
  get observable(): Observable<Session> {
    return this._subject
  }
  get ready(): Promise<boolean> {
    return this._ready.promise
  }
  private async _restore() {
    console.log('start session restore')
    await this._mutex.lock()
    try {
      const response = await ServerAPI.v1.session.get<Session>(this.httpClient)
      if (response && typeof response.name === "string") {
        console.info(`session restore`, response)
        this._subject.next(response)
      }
    } catch (e) {
      console.error(`restore error : `, e)
    } finally {
      this._mutex.unlock()
      this._ready.resolve(true)
    }
  }
  /**
  * 登入
  * @param name 
  * @param password 
  * @param keep 
  */
  async login(name: string, password: string, remember: boolean): Promise<Session> {
    await this._mutex.lock()
    let result: Session
    try {
      const response = await ServerAPI.v1.session.post<Session>(this.httpClient, {
        name: name,
        password: password,
        remember: remember,
      })
      if (response) {
        console.info(`login success`, response)
        this._subject.next(response)
      } else {
        console.warn(`login unknow result`, response)
        throw new Exception("login unknow result")
      }
    } finally {
      this._mutex.unlock()
    }
    return result
  }

  /**
   * 登出
   */
  async logout() {
    await this._mutex.lock()
    try {
      if (this._subject.value == null) {
        return
      }
      await ServerAPI.v1.session.delete(this.httpClient)
      this._subject.next(null)
    } finally {
      this._mutex.unlock()
    }
  }
}
