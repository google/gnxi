import { Directive, HostBinding } from '@angular/core';

@Directive({
  selector: '[appFileUpload]'
})
export class FileUploadDirective {

  @HostBinding('class.fileover') fileOver: boolean;

  constructor() { }

}
