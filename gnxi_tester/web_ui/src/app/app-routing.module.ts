import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { RunComponent } from './run/run.component';
import { DevicesComponent } from './devices/devices.component';
import { PromptsComponent } from './prompts/prompts.component';

const routes: Routes = [
  {path: '', component: RunComponent},
  {path: 'devices', component: DevicesComponent},
  {path: 'prompts', component: PromptsComponent},
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
