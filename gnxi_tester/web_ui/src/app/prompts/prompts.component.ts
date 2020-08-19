import { Component, OnInit } from '@angular/core';
import { PromptsService } from '../prompts.service';
import { FormControl } from '@angular/forms';
import { PromptsList, PromptsSet } from '../models/Prompts';

type ControlGroup = {[name: string]: FormControl}

@Component({
  selector: 'app-prompts',
  templateUrl: './prompts.component.html',
  styleUrls: ['./prompts.component.css']
})
export class PromptsComponent implements OnInit {

  constructor(public promptsService: PromptsService) {
    this.init()
  }

  async ngOnInit() {
  }

  async init() {
    try {
      this.promptsList = await this.promptsService.getPromptsList().toPromise()
      this.prompts = await this.promptsService.getPrompts().toPromise()
      this.promptsNames = Object.keys(this.prompts)
      for (let field of this.promptsList.prompts) {
        this.controlGroupPrompts[field] = new FormControl("");
      }
      for (let field of this.promptsList.files) {
        this.controlGroupFiles[field] = new FormControl("");
      }
    } catch(e) {
      console.error(e)
    }
  }

  promptsList: PromptsList
  promptsNames: string[];

  controlGroupPrompts: ControlGroup = {};
  controlGroupFiles: ControlGroup = {};
  prompts: PromptsSet;
  selected = new FormControl("");
}
