import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from './environment';
import { FileResponse } from './models/FileResponse';

@Injectable({
  providedIn: 'root'
})
export class FileService {

  constructor(private http: HttpClient) { }

  deleteFile(filename: string) {
    return this.http.delete(`${environment.apiUrl}/file/${filename}`);
  }

  uploadFile(file: File): Observable<FileResponse> {
    const uploadForm = new FormData();
    uploadForm.set('file', file);
    return this.http.post<FileResponse>(`${environment.apiUrl}/file`, uploadForm);
  }
}
