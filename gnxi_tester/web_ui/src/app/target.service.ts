import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Target, Targets } from './models/Target';
import { environment } from './environment';

@Injectable({
  providedIn: 'root'
})
export class TargetService {

  constructor(private http: HttpClient) { }

  getTargets(): Observable<Targets> {
    return this.http.get<Targets>(`${environment.apiUrl}/target`);
  }

  getTarget(name: string): Observable<Target> {
    return this.http.get<Target>(`${environment.apiUrl}/target/${name}`);
  }
  delete(name: string) {
    return this.http.delete(`${environment.apiUrl}/target/${name}`);
  }
}
