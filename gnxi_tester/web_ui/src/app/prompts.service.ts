import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from './environment';
import { Prompts, PromptsList, PromptsSet } from './models/Prompts';


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

  delete(name: string) {
    return this.http.delete(`${environment.apiUrl}/prompts/${name}`);
  }
}
