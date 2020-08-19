import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs'
import { Target, Targets } from './models/Target';

@Injectable({
  providedIn: 'root'
})
export class TargetService {

  constructor(private http: HttpClient) { }

  getTargets(): Observable<Targets> {
    return this.http.get<Targets>('http://localhost:8888/target');
  }

  getTarget(name: string): Observable<Target> {
    return this.http.get<Target>(`http://localhost:8888/target/${name}`);
  }
}
