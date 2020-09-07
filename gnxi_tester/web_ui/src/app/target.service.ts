import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from './environment';
import { Target, Targets } from './models/Target';

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
