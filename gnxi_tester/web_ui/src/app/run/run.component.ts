import { CdkDragDrop, moveItemInArray } from '@angular/cdk/drag-drop';
import { Component, ElementRef, OnInit, ViewChild } from '@angular/core';
import { AbstractControl, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { PromptsSet } from '../models/Prompts';
import { Targets } from '../models/Target';
import { PromptsService } from '../prompts.service';
import { RunService } from '../run.service';
import { TargetService } from '../target.service';
import { TestService } from '../test.service';

@Component({
  selector: 'app-run',
  templateUrl: './run.component.html',
  styleUrls: ['./run.component.css']
})
export class RunComponent implements OnInit {
  promptsList: PromptsSet;
  deviceList: Targets;
  runForm: FormGroup;
  testNames: string[];
  sample = 'Test output will go here';
  stdout = this.sample;
  @ViewChild('terminal') private terminal: ElementRef<HTMLDivElement>;

  constructor(public runService: RunService, public promptsService: PromptsService, public targetService: TargetService, private formBuilder: FormBuilder, public testService: TestService) { }

  async ngOnInit() {
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
      tests: [[]],
    });
    let output = await this.runService.runOutput().toPromise();
    if (output) {
      this.runOutput();
    }
    this.testService.getTestOrder().subscribe(
      (testNames) => this.testNames = testNames,
      (err) => console.log(err)
    );
  }

  run(runForm: any): void {
    this.runService.run(runForm).subscribe((res) => {
      console.log(res);
    }, error => console.error(error));
    this.runOutput();
  }

  private runOutput(): void {
    this.runForm.disable();
    this.stdout = '';
    let inter = setInterval(async () => {
      let added = await this.runService.runOutput().toPromise();
      if (added === null) {
        return;
      }
      if (added === 'E0F') {
        this.runForm.enable();
        clearInterval(inter);
      } else {
        this.stdout += added;
      }
      this.terminal.nativeElement.scrollTop = this.terminal.nativeElement.scrollHeight;
    }, 1000);
  }

  dropTestChip(event: CdkDragDrop<string[]>): void {
    moveItemInArray(this.testNames, event.previousIndex, event.currentIndex);
    this.runForm.patchValue({ tests: this.testNames });
  }

  removeTest(name: string): void {
    this.testNames = this.testNames.filter((value, index) => {
      return value !== name;
    });
    this.runForm.patchValue({ tests: this.testNames });
  }

  get device(): AbstractControl {
    return this.runForm.get('device');
  }


  get prompts(): AbstractControl {
    return this.runForm.get('prompts');
  }
}
