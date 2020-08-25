import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { RunRequest } from './models/Run';
import { Observable } from 'rxjs';
import { environment } from './environment';

@Injectable({
  providedIn: 'root'
})
export class RunService {

  constructor(private http: HttpClient) { }

  run(req: RunRequest) {
    this.http.post(`${environment.apiUrl}/run`, req)
  }
  runOutput(): Observable<string> {
    return this.http.get<string>(`${environment.apiUrl}/run/output`)
  }
}
