import { Component, OnInit } from '@angular/core';
import { PromptsService } from '../prompts.service';
import { FormControl, Validators, FormBuilder, FormGroup } from '@angular/forms';
import { PromptsList, PromptsSet, Prompts } from '../models/Prompts';

type ControlGroup = {[name: string]: FormControl}

@Component({
  selector: 'app-prompts',
  templateUrl: './prompts.component.html',
  styleUrls: ['./prompts.component.css']
})
export class PromptsComponent implements OnInit {

  constructor(public promptsService: PromptsService, private formBuilder: FormBuilder) {
    this.init()
  }

  async ngOnInit() {
  }

  async init() {
    try {
      this.promptsList = await this.promptsService.getPromptsList().toPromise()
      this.prompts = await this.promptsService.getPrompts().toPromise()
      let fields = this.getFields();
      this.controlGroup = this.formBuilder.group(fields)
    } catch(e) {
      console.error(e)
    }
  }


  private getFields(): {[name: string]: any} {
    let fields = {name: ['',Validators.required]};
    for (let field of this.promptsList.prompts) {
      fields["prompts_"+field] = ['', Validators.required]
    }
    for (let field of this.promptsList.files) {
      fields["files_"+field] = ['']
    }
    return fields
  }

  promptsList: PromptsList = {prompts: [], files: []};

  controlGroup = this.formBuilder.group({name: ['', Validators.required]});
  prompts: PromptsSet;

  setFile(name: string, val: string) {
    let fields = {};
    fields[`files_${name}`] = val;
    this.controlGroup.patchValue(fields)
  }

  setPrompts(form: {[key: string]: string}): void {
    let prompts: Prompts = {
      name: form.name,
      prompts: {},
      files: {}
    }
    for (let field of Object.keys(form)) {
      if (field.search("prompts_")) {
        let key = field.slice(8);
        prompts.prompts[key] = form[field];
      } else if (field.search("files_")) {
        let key = field.slice(6);
        prompts.files[key] = form[field];
      }
    }
    this.promptsService.setPrompts(prompts).subscribe(
      (res) => console.log(res),
      (err) => console.log(err)
    );
  }

  setSelectedPrompts(name: string): void {
    const prompts = this.prompts[name];
    if (prompts === undefined) {
      this.controlGroup.reset();
      return;
    }
    let fields = {};
    for (let field of Object.keys(prompts.files)) {
      fields[`files_${field}`] = prompts.files[field];
    }
    for (let field of Object.keys(prompts.prompts)) {
      fields[`prompts_${field}`] = prompts.prompts[field];
    }
    this.controlGroup.setValue({
      name,
      ...fields,
    });
  }
}
