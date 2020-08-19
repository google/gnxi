import { Component, OnInit } from '@angular/core';
import { PromptsService, Prompts } from '../prompts.service';
import { FormControl } from '@angular/forms';

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
    } catch(e) {
      console.error(e)
    }
  }

  promptsList: string[]
  selected = new FormControl("");
}
