import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { NavbarComponent } from './navbar/navbar.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { MatButtonModule } from '@angular/material/button';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatCardModule } from '@angular/material/card';
import { MatGridListModule } from '@angular/material/grid-list';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { ReactiveFormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { RunComponent } from './run/run.component';
import { PromptsComponent } from './prompts/prompts.component';
import { DevicesComponent } from './devices/devices.component';
import { FileUploadComponent } from './file-upload/file-upload.component';
import { FileUploadDirective } from './file-upload.directive';

@NgModule({
  declarations: [
    AppComponent,
    NavbarComponent,
    RunComponent,
    PromptsComponent,
    DevicesComponent,
    FileUploadComponent,
    FileUploadDirective
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    MatButtonModule,
    MatToolbarModule,
    HttpClientModule,
    MatChipsModule,
    BrowserAnimationsModule,
    MatSelectModule,
    FormsModule,
    MatFormFieldModule,
    MatProgressSpinnerModule,
    MatCardModule,
    MatGridListModule,
    MatIconModule,
    MatInputModule,
    ReactiveFormsModule,
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
