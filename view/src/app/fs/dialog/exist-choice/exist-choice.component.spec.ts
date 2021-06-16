import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';

import { ExistChoiceComponent } from './exist-choice.component';

describe('ExistChoiceComponent', () => {
  let component: ExistChoiceComponent;
  let fixture: ComponentFixture<ExistChoiceComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [ ExistChoiceComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ExistChoiceComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
