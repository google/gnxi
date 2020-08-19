import { Component, OnInit, Input, HostListener, Output, EventEmitter } from '@angular/core';

@Component({
  selector: 'app-file-upload',
  templateUrl: './file-upload.component.html',
  styleUrls: ['./file-upload.component.css']
})
export class FileUploadComponent implements OnInit {

  constructor() { }

  ngOnInit(): void {
  }

  @Input() name: string;
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
      const returnedID = "";
      this.ID = returnedID;
      this.getID.emit(this.ID);
      console.log(files[0]);
      this.progress = false;
    }
  }

  fileOver = false;
}
