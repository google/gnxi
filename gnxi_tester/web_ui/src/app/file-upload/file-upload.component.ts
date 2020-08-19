import { Component, OnInit, Input, HostListener, Output, EventEmitter } from '@angular/core';
import { FileService } from '../file.service';

@Component({
  selector: 'app-file-upload',
  templateUrl: './file-upload.component.html',
  styleUrls: ['./file-upload.component.css']
})
export class FileUploadComponent implements OnInit {

  constructor(private fileService: FileService) { }

  ngOnInit(): void {
  }

  @Input() name: string;
  @Input() path: string;
  @Input() progress: boolean;
  @Input() ID: string;
  @Output() getID = new EventEmitter<string>();

  @HostListener('dragover', ['$event']) onDragOver(e: DragEvent) {
    e.preventDefault();
    e.stopPropagation();
    console.log("over");
    this.fileOver = true;
  }

  @HostListener('dragleave', ['$event']) onDragLeave(e: DragEvent) {
    e.preventDefault();
    e.stopPropagation();
    console.log("leave");
  }

  @HostListener('drop', ['$event']) onDrop(e: DragEvent) {
    e.preventDefault();
    e.stopPropagation();
    console.log("drop");
    this.fileOver = false;
    this.upload(e.dataTransfer.files);
  }

  @HostListener('change', ['$event']) onUpload(e: any) {
    if(e?.target?.files) {
      this.upload(e.target.files);
    }

  }
  private upload(files: FileList) {
    if (files?.length) {
      this.progress = true;
      // Upload the file.
      this.fileService.uploadFile(files[0]).subscribe((res) => {
          this.ID = res.filename;
          this.getID.emit(this.ID);
          this.progress = false;
        }
      );
    }
  }

  fileOver = false;
}
