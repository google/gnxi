import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { TargetsComponents } from './targets.component';

describe('TargetsComponents', () => {
  let component: TargetsComponents;
  let fixture: ComponentFixture<TargetsComponents>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ TargetsComponents ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(TargetsComponents);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
