import { TestBed } from '@angular/core/testing';

import { PromptsService } from './prompts.service';

describe('PromptsService', () => {
  let service: PromptsService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(PromptsService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
