import { HttpClient } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { AbstractControl, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatSnackBar } from '@angular/material/snack-bar';
import { environment } from '../environment';
import { FileService } from '../file.service';
import { Targets } from '../models/Target';
import { TargetService } from '../target.service';

@Component({
  selector: 'app-devices',
  templateUrl: './devices.component.html',
  styleUrls: ['./devices.component.css']
})
export class DevicesComponent implements OnInit {
  targetList: Targets = {};
  targetForm: FormGroup;
  selectedTarget: any = {};

  constructor(private http: HttpClient, private targetService: TargetService, private formBuilder: FormBuilder, private fileService: FileService, private snackbar: MatSnackBar) { }

  ngOnInit(): void {
    this.targetForm = this.formBuilder.group({
      targetName: ['', Validators.required],
      address: ['', Validators.required],
      ca: ['', Validators.required],
      cakey: ['', Validators.required],
    });
    this.getTargets();
  }

  getTargets(): void {
    this.targetService.getTargets().subscribe(targets => {
      this.targetList = targets;
    });
  }

  setTarget(targetForm): void {
    this.http.post(`${environment.apiUrl}/target/${targetForm.targetName}`, targetForm).subscribe(
      (res) => {
        this.targetList[targetForm.targetName] = {
        address: targetForm.address,
        ca: targetForm.ca,
        cakey: targetForm.cakey,
      };
        this.snackbar.open("Saved", "", {duration: 2000});
    },
      (error) => console.error(error),
    );
  }

  deleteTarget(): void {
    let name = this.targetForm.get("targetName").value;
    this.targetService.delete(name).subscribe(res => {
      this.selectedTarget = {};
      this.targetForm.reset();
      delete this.targetList[name];
      this.snackbar.open("Deleted", "", {duration: 2000});
    }, error => console.error(error))
  }

  setSelectedTarget(targetName: string): void {
    this.selectedTarget = this.targetList[targetName];
    if (this.selectedTarget === undefined) {
      this.selectedTarget = {};
      this.targetForm.reset();
      return;
    }
    this.targetForm.setValue({
      targetName,
      address: this.selectedTarget.address,
      ca: this.selectedTarget.ca,
      cakey: this.selectedTarget.cakey,
    });
  }

  addCa(caFileName: string): void {
    this.targetForm.patchValue({
       ca: caFileName,
    });
  }

  addCaKey(keyFileName: string): void {
    this.targetForm.patchValue({
      cakey: keyFileName,
    });
  }

  get targetName(): AbstractControl {
    return this.targetForm.get('targetName');
  }
  get targetAddress(): AbstractControl {
    return this.targetForm.get('address');
  }
}
