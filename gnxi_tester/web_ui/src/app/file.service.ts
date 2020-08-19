import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { FileResponse } from './models/FileResponse';

@Injectable({
  providedIn: 'root'
})
export class FileService {

  constructor(private http: HttpClient) { }

  deleteFile(filename: string) {
    return this.http.delete(`http://localhost:8888/file/${filename}`);
  }

  uploadFile(file: File): Observable<FileResponse> {
    const uploadForm = new FormData();
    uploadForm.set('file', file);
    return this.http.post<FileResponse>('http://localhost:8888/file', uploadForm);
  }
}
