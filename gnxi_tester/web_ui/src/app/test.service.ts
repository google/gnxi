import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from './environment';
import { Tests } from './models/Test';

@Injectable({
  providedIn: 'root'
})
export class TestService {

  constructor(private http: HttpClient) { }

  getTests(): Observable<Tests> {
    return this.http.get<Tests>(`${environment.apiUrl}/test`)
  }
}
