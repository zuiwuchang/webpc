import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ExistChoiceComponent } from './exist-choice.component';

describe('ExistChoiceComponent', () => {
  let component: ExistChoiceComponent;
  let fixture: ComponentFixture<ExistChoiceComponent>;

  beforeEach(async(() => {
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
