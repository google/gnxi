import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class FileService {

  constructor(private http: HttpClient) { }

  deleteFile(filename: string): void {
    this.http.delete(`http://localhost:8888/file/${filename}`).subscribe(res => {
      console.log(res);
    });
  }

  uploadFile(file: File): Observable<string> {
    const uploadForm = new FormData();
    uploadForm.set('file', file);
    return this.http.post<string>('http://localhost:8888/file', uploadForm);
  }
}
