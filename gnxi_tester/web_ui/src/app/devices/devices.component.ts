import { Component, OnInit } from '@angular/core';
import { FormGroup, Validators, FormBuilder, AbstractControl } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { TargetService } from '../target.service';
import { FileService } from '../file.service';
import { Targets } from '../models/Target';

@Component({
  selector: 'app-devices',
  templateUrl: './devices.component.html',
  styleUrls: ['./devices.component.css']
})
export class DevicesComponent implements OnInit {
  targetList: Targets;
  targetForm: FormGroup;

  constructor(private http: HttpClient, private targetService: TargetService, private formBuilder: FormBuilder, private fileService: FileService) { }

  ngOnInit(): void {
    this.targetForm = this.formBuilder.group({
      targetName: ['', Validators.required],
      targetAddress: ['', Validators.required],
      caCert: '',
      caKey: '',
    });
    this.getTargets();
  }

  getTargets(): void {
    this.targetService.getTargets().subscribe(targets => {
      this.targetList = targets;
    });
  }

  setTarget(targetForm): void {
    this.http.post(`http://localhost:8888/target/${targetForm.targetName}`, targetForm).subscribe((res) => {
      console.log(res);
    });
  }

  setSelectedTarget(targetName: string): void {
    const target = this.targetList[targetName];
    if (target === undefined) {
      this.targetForm.reset();
      return;
    }
    this.targetForm.setValue({
      targetName,
      targetAddress: target.address,
      caCert: target.ca,
      caKey: target.cakey,
    });
  }

  addCa(caFileName: string): void {
    this.targetForm.patchValue({
       caCert: caFileName,
    });
  }

  addCaKey(keyFileName: string): void {
    this.targetForm.patchValue({
      caKey: keyFileName,
    });
  }

  get targetName(): AbstractControl {
    return this.targetForm.get('targetName');
  }
  get targetAddress(): AbstractControl {
    return this.targetForm.get('targetAddress');
  }
}
