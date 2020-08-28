import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormControl, Validators, FormGroup } from '@angular/forms';
import { MatSnackBar } from '@angular/material/snack-bar';
import { FileService } from '../file.service';
import { Prompts, PromptsList, PromptsSet } from '../models/Prompts';
import { PromptsService } from '../prompts.service';

@Component({
  selector: 'app-prompts',
  templateUrl: './prompts.component.html',
  styleUrls: ['./prompts.component.css']
})
export class PromptsComponent implements OnInit {
  prompts: PromptsSet = {};
  promptsList: PromptsList = { prompts: [], files: [] };
  promptsForm: FormGroup;
  files = {};

  constructor(public promptsService: PromptsService, private formBuilder: FormBuilder, private snackbar: MatSnackBar, private fileService: FileService) {}

  ngOnInit(): void {
    this.promptsService.getPromptsList().subscribe(
      (promptsList) => {
        this.promptsList = promptsList;
        this.promptsForm = this.createForm();
      },
      (err) => console.log(err)
    );
    this.promptsService.getPrompts().subscribe(
      (prompts) => this.prompts = prompts,
      (err) => console.log(err)
    );
  }

  private createForm(): FormGroup {
    let fields = { name: new FormControl('', Validators.required) };
    for (let field of this.promptsList.prompts) {
      fields['prompts_' + field] = new FormControl('', Validators.required);
    }
    for (let field of this.promptsList.files) {
      fields['files_' + field] = new FormControl('');
    }
    return new FormGroup(fields);
  }

  setFile(name: string, val: string) {
    let fields = {};
    fields[`files_${name}`] = val;
    this.promptsForm.patchValue(fields)
  }

  deletePrompts(): void {
    let name = this.promptsForm.get('name').value;
    this.promptsService.delete(name).subscribe(res => {
      console.log(res);
      delete this.prompts[name];
      this.promptsForm.reset()
      this.snackbar.open('Deleted', '', {duration: 2000});
    }, error => console.error(error))
  }

  setPrompts(form: { [key: string]: string }): void {
    if (!form) {
      return;
    }
    console.log(form);
    let prompts: Prompts = {
      name: form.name,
      prompts: {},
      files: {}
    }
    for (let field of Object.keys(form)) {
      if (field.search('prompts_') === 0) {
        let key = field.slice(8);
        prompts.prompts[key] = form[field];
      } else if (field.search('files_') === 0) {
        let key = field.slice(6);
        prompts.files[key] = form[field];
      }
    }
    this.promptsService.setPrompts(prompts).subscribe(
      (res) => {
        console.log(res);
        this.snackbar.open('Saved', '', { duration: 2000 });
        this.prompts[prompts.name] = prompts;
        this.setSelectedPrompts(prompts.name);
      },
      (err) => console.log(err)
    );
  }

  setSelectedPrompts(name: string): void {
    const prompts = this.prompts[name];
    if (prompts === undefined) {
      this.promptsForm.reset();
      return;
    }
    let fields = {};
    console.log(prompts.files)
    for (let field of Object.keys(prompts.files)) {
      fields[`files_${field}`] = prompts.files[field];
    }
    for (let field of Object.keys(prompts.prompts)) {
      fields[`prompts_${field}`] = prompts.prompts[field];
    }
    this.promptsForm.setValue({
      name,
      ...fields,
    });
  }

  get selected() {
    return this.promptsForm.get('name').value;
  }
}
