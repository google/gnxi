import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Prompts {
  name: string;
  prompts: {[name: string]: string};
  files: {[name: string]: string};
}

export interface PromptsList {
  prompts: string[];
  files: string[];
}

export type PromptsSet = {[name: string]: Prompts}

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
    this.http.post("http://localhost:8888/prompts", prompts)
  }
}
