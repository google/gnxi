import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { RunComponent } from './run/run.component';
import { TargetsComponent } from './targets/targets.component';
import { PromptsComponent } from './prompts/prompts.component';

const routes: Routes = [
  {path: '', redirectTo: 'run', pathMatch: 'prefix'},
  {path: 'run', component: RunComponent},
  {path: 'targets', component: TargetsComponent},
  {path: 'prompts', component: PromptsComponent},
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
