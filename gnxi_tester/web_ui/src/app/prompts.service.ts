import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { PromptsSet, PromptsList, Prompts } from './models/Prompts';

const httpOptions = {
  headers: new HttpHeaders({
    'Content-Type':  'application/json',
  })
};

@Injectable({
  providedIn: 'root'
})
export class PromptsService {
  constructor(private http: HttpClient) {
  }

  getPrompts(): Observable<PromptsSet> {
    return this.http.get<PromptsSet>("http://localhost:8888/prompts")
  }

  getPromptsList(): Observable<PromptsList> {
    return this.http.get<PromptsList>("http://localhost:8888/prompts/list")
  }

  setPrompts(prompts: Prompts) {
    try {
      this.http.post<Prompts>("http://localhost:8888/prompts", prompts, httpOptions).subscribe(res => console.log(res));
    } catch (e) {
      console.error(e);
    }
  }
}
