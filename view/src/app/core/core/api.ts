import { RESTful } from './restful';
import { environment } from 'src/environments/environment';
const root = '/api'

export const ServerAPI = {
    v1: {
        version: new RESTful(root, 'v1', 'version'),
        roots: new RESTful(root, 'v1', 'roots'),
        debug: new RESTful(root, 'v1', 'debug'),

        session: new RESTful(root, 'v1', 'session'),
        users: new RESTful(root, 'v1', 'users'),
        shells: new RESTful(root, 'v1', 'shells'),

        fs: new RESTful(root, 'v1', 'fs'),
    },
}