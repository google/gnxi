import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Prompts {
  name: string;
  prompts: {[name: string]: string};
  files: {[name: string]: string};
}

@Injectable({
  providedIn: 'root'
})
export class PromptsService {
  constructor(private http: HttpClient) {
  }

  getPrompts(): Observable<Prompts[]> {
    return this.http.get<Prompts[]>("http://localhost:8888/prompts")
  }

  getPromptsList(): Observable<string[]> {
    return this.http.get<string[]>("http://localhost:8888/prompts/list")
  }

  setPrompts(prompts: Prompts) {
    this.http.post("http://localhost:8888/prompts", prompts)
  }
}
