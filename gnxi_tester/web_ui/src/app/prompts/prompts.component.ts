import { Component, OnInit } from '@angular/core';
import { PromptsService, Prompts, PromptsSet, PromptsList } from '../prompts.service';
import { FormControl } from '@angular/forms';

type ControlGroup = {[name: string]: FormControl}

@Component({
  selector: 'app-prompts',
  templateUrl: './prompts.component.html',
  styleUrls: ['./prompts.component.css']
})
export class PromptsComponent implements OnInit {

  constructor(public promptsService: PromptsService) {}

  async ngOnInit() {
    try {
      this.promptsList = await this.promptsService.getPromptsList().toPromise()
      this.prompts = await this.promptsService.getPrompts().toPromise()
      this.promptsNames = Object.keys(this.prompts)
      for (let field of this.promptsList.prompts) {
        this.controlGroup[field] = new FormControl("");
      }
    } catch(e) {
      console.error(e)
    }
  }

  promptsList: PromptsList
  promptsNames: string[];

  controlGroup: ControlGroup = {};
  prompts: PromptsSet;
  selected = new FormControl("");
}
