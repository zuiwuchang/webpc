import { HttpHeaders, HttpParams, HttpClient } from '@angular/common/http'
import { isNumber, isString, isObject } from 'util'

export function resolveError(e): string {
    if (!e) {
        return "nil"
    }
    if (isString(e)) {
        return e
    }
    if (isObject(e) && isNumber(e.status)) {
        return resolveHttpError(e)
    }
    return "unknow"
}
export function resolveHttpError(e) {
    if (isString(e.error)) {
        return `${e.status} ${e.error}`
    }
    if (e.error) {
        return `${e.status} ${e.error.description}`
    } else if (e.message) {
        return `${e.status} ${e.message}`
    }
    return `${e.status} ${e.statusText}`
}
export function wrapPromise<T>(promise: Promise<T>): Promise<T> {
    return new Promise<T>(function (resolve, reject) {
        promise.then(
            function (v) {
                resolve(v)
            },
            function (e) {
                reject(resolveError(e))
            },
        )
    })
}
export class RESTful {
    constructor(public baseURL: string) {

    }
    oneURL(id: string | number | boolean): string {
        return `${this.baseURL}/${encodeURIComponent(encodeURIComponent(id))}`
    }
    onePatchURL(id: string | number | boolean, patch: string): string {
        return `${this.oneURL(id)}/${patch}`
    }
    get<T>(client: HttpClient, options?: {
        headers?: HttpHeaders | {
            [header: string]: string | string[];
        };
        observe?: 'body';
        params?: HttpParams | {
            [param: string]: string | string[];
        };
        reportProgress?: boolean;
        responseType?: 'json';
        withCredentials?: boolean;
    }): Promise<T> {
        return wrapPromise(client.get<T>(this.baseURL, options).toPromise())
    }
    post<T>(client: HttpClient, body: any | null, options?: {
        headers?: HttpHeaders | {
            [header: string]: string | string[];
        };
        observe?: 'body';
        params?: HttpParams | {
            [param: string]: string | string[];
        };
        reportProgress?: boolean;
        responseType?: 'json';
        withCredentials?: boolean;
    }): Promise<T> {
        return wrapPromise(client.post<T>(this.baseURL, body, options).toPromise())
    }
    delete<T>(client: HttpClient, options?: {
        headers?: HttpHeaders | {
            [header: string]: string | string[];
        };
        observe?: 'body';
        params?: HttpParams | {
            [param: string]: string | string[];
        };
        reportProgress?: boolean;
        responseType?: 'json';
        withCredentials?: boolean;
    }): Promise<T> {
        return wrapPromise(client.delete<T>(this.baseURL, options).toPromise())
    }
    put<T>(client: HttpClient, body: any | null, options?: {
        headers?: HttpHeaders | {
            [header: string]: string | string[];
        };
        observe?: 'body';
        params?: HttpParams | {
            [param: string]: string | string[];
        };
        reportProgress?: boolean;
        responseType?: 'json';
        withCredentials?: boolean;
    }): Promise<T> {
        return wrapPromise(client.put<T>(this.baseURL, body, options).toPromise())
    }
    patch<T>(client: HttpClient, patch: string, body: any | null, options?: {
        headers?: HttpHeaders | {
            [header: string]: string | string[];
        };
        observe?: 'body';
        params?: HttpParams | {
            [param: string]: string | string[];
        };
        reportProgress?: boolean;
        responseType?: 'json';
        withCredentials?: boolean;
    }): Promise<T> {
        return wrapPromise(client.patch<T>(`${this.baseURL}/${patch}`, body, options).toPromise())
    }

    getOne<T>(client: HttpClient, id: string | number | boolean, options?: {
        headers?: HttpHeaders | {
            [header: string]: string | string[];
        };
        observe?: 'body';
        params?: HttpParams | {
            [param: string]: string | string[];
        };
        reportProgress?: boolean;
        responseType?: 'json';
        withCredentials?: boolean;
    }): Promise<T> {
        return wrapPromise(client.get<T>(this.oneURL(id), options).toPromise())
    }
    postOne<T>(client: HttpClient, id: string | number | boolean, body: any | null, options?: {
        headers?: HttpHeaders | {
            [header: string]: string | string[];
        };
        observe?: 'body';
        params?: HttpParams | {
            [param: string]: string | string[];
        };
        reportProgress?: boolean;
        responseType?: 'json';
        withCredentials?: boolean;
    }): Promise<T> {
        return wrapPromise(client.post<T>(this.oneURL(id), body, options).toPromise())
    }
    deleteOne<T>(client: HttpClient, id: string | number | boolean, options?: {
        headers?: HttpHeaders | {
            [header: string]: string | string[];
        };
        observe?: 'body';
        params?: HttpParams | {
            [param: string]: string | string[];
        };
        reportProgress?: boolean;
        responseType?: 'json';
        withCredentials?: boolean;
    }): Promise<T> {
        return wrapPromise(client.delete<T>(this.oneURL(id), options).toPromise())
    }
    putOne<T>(client: HttpClient, id: string | number | boolean, body: any | null, options?: {
        headers?: HttpHeaders | {
            [header: string]: string | string[];
        };
        observe?: 'body';
        params?: HttpParams | {
            [param: string]: string | string[];
        };
        reportProgress?: boolean;
        responseType?: 'json';
        withCredentials?: boolean;
    }): Promise<T> {
        return wrapPromise(client.put<T>(this.oneURL(id), body, options).toPromise())
    }
    patchOne<T>(client: HttpClient, id: string | number | boolean, patch: string, body: any | null, options?: {
        headers?: HttpHeaders | {
            [header: string]: string | string[];
        };
        observe?: 'body';
        params?: HttpParams | {
            [param: string]: string | string[];
        };
        reportProgress?: boolean;
        responseType?: 'json';
        withCredentials?: boolean;
    }): Promise<T> {
        return wrapPromise(client.patch<T>(this.onePatchURL(id, patch), body, options).toPromise())
    }
}