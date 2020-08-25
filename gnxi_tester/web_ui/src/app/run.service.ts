import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { RunRequest } from './models/Run';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class RunService {

  constructor(private http: HttpClient) { }

  run(req: RunRequest) {
    this.http.post("http://localhost:8888/run", req)
  }
  runOutput(): Observable<string> {
    return this.http.get<string>("http://localhost:8888/run/output")
  }
}
