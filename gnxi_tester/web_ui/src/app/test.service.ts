import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from './environment';
import { Test } from './models/Test';

@Injectable({
  providedIn: 'root'
})
export class TestService {

  constructor(private http: HttpClient) { }

  getTests(): Observable<Map<string, Test[]>> {
    return this.http.get<Map<string, Test[]>>(`${environment.apiUrl}/test`);
  }

  getTestOrder(): Observable<string[]> {
    return this.http.get<string[]>(`${environment.apiUrl}/test/order`);
  }
}
