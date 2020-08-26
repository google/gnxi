import { Component, OnInit } from '@angular/core';
import { AbstractControl, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { PromptsSet } from '../models/Prompts';
import { Targets } from '../models/Target';
import { PromptsService } from '../prompts.service';
import { RunService } from '../run.service';
import { TargetService } from '../target.service';

@Component({
  selector: 'app-run',
  templateUrl: './run.component.html',
  styleUrls: ['./run.component.css']
})
export class RunComponent implements OnInit {
  promptsList: PromptsSet;
  deviceList: Targets;
  runForm: FormGroup;
  sample = 'Test output will go here';
  stdout = this.sample;

  constructor(public runService: RunService, public promptsService: PromptsService, public targetService: TargetService, private formBuilder: FormBuilder) {}

  ngOnInit(): void {
    this.targetService.getTargets().subscribe(
      (res) => this.deviceList = res,
      (err) => console.log(err)
    );
    this.promptsService.getPrompts().subscribe(
      (res) => this.promptsList = res,
      (err) => console.log(err)
    );
    this.runForm = this.formBuilder.group({
      prompts: ['', Validators.required],
      device: ['', Validators.required],
    });
  }

  run(runForm: any): void {
    this.runService.run(runForm); // I think a subscribe should go here?
    this.runForm.disable();
    this.stdout = '';
    let inter = setInterval(async () => {
        let added = await this.runService.runOutput().toPromise();
        if (added === 'E0F') {
          this.runForm.enable();
          clearInterval(inter);
        } else {
          this.stdout += added;
        }
    }, 1)
  }

  get device(): AbstractControl {
    return this.runForm.get('device');
  }


  get prompts(): AbstractControl {
    return this.runForm.get('prompts');
  }
}
