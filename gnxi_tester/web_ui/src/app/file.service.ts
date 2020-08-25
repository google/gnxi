import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { FileResponse } from './models/FileResponse';
import { environment } from './environment';

@Injectable({
  providedIn: 'root'
})
export class FileService {

  constructor(private http: HttpClient) { }

  deleteFile(filename: string) {
    return this.http.delete(`${environment.apiUrl}/${filename}`);
  }

  uploadFile(file: File): Observable<FileResponse> {
    const uploadForm = new FormData();
    uploadForm.set('file', file);
    return this.http.post<FileResponse>(`${environment.apiUrl}/file`, uploadForm);
  }
}
