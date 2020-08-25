import { Component, OnInit } from '@angular/core';
import { PromptsService } from '../prompts.service';
import { TargetService } from '../target.service';
import { RunService } from '../run.service';
import { FormBuilder, Validators } from '@angular/forms';
import { RunRequest } from '../models/Run';

@Component({
  selector: 'app-run',
  templateUrl: './run.component.html',
  styleUrls: ['./run.component.css']
})
export class RunComponent implements OnInit {

  constructor(public runService: RunService, public promptsService: PromptsService, public targetService: TargetService, private formBuilder: FormBuilder) {}

  ngOnInit(): void {
  }

  async init() {
    try {
      let promptsSet = await this.promptsService.getPrompts().toPromise()
      this.promptsList = Object.keys(promptsSet)
      let deviceList = await this.targetService.getTargets().toPromise()
      this.deviceList = Object.keys(deviceList)
    } catch(e) {
      console.error(e)
    }
  }

  run(form: {[name: string]: string}) {
    this.runService.run(form as any as RunRequest);
    this.running = true;
    this.stdout = "";
    let inter = setInterval(async () => {
        let added = await this.runService.runOutput().toPromise();
        if (added === 'E0F') {
          this.running = false;
          clearInterval(inter);
        } else {
          this.stdout += added;
        }
    }, 1)
  }

  running = false;
  sample = "Test output will go here";
  stdout = this.sample;

  promptsList: string[] = [];

  formControl = this.formBuilder.group({
    prompts: ["", Validators.required],
    device: ["", Validators.required]
  })
  deviceList: string[] = [];
}
