import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { CutComponent } from './cut.component';

describe('CutComponent', () => {
  let component: CutComponent;
  let fixture: ComponentFixture<CutComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ CutComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(CutComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
