import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-run',
  templateUrl: './run.component.html',
  styleUrls: ['./run.component.css']
})
export class RunComponent implements OnInit {

  constructor() {}

  ngOnInit(): void {
  }

  selectedPrompts: string;
  promptsList: string[] = ["test", "test2"];

  selectedDevice: string;
  deviceList: string[] = ["test", "test2"];
}
