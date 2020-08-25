import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { PromptsSet, PromptsList, Prompts } from './models/Prompts';
import { environment } from './environment';


@Injectable({
  providedIn: 'root'
})
export class PromptsService {
  constructor(private http: HttpClient) {
  }

  getPrompts(): Observable<PromptsSet> {
    return this.http.get<PromptsSet>(`${environment.apiUrl}/prompts`);
  }

  getPromptsList(): Observable<PromptsList> {
    return this.http.get<PromptsList>(`${environment.apiUrl}/prompts/list`);
  }

  setPrompts(prompts: Prompts): any {
    return this.http.post(`${environment.apiUrl}/prompts`, prompts);
  }
}
