import { Component, OnInit, Input, HostListener } from '@angular/core';

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
    const files = e.dataTransfer.files;
    if (files.length) {
      // Upload the file.
      console.log(files[0]);
    }
  }
  fileOver = false;
}
