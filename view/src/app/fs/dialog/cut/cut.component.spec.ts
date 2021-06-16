import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';

import { CutComponent } from './cut.component';

describe('CutComponent', () => {
  let component: CutComponent;
  let fixture: ComponentFixture<CutComponent>;

  beforeEach(waitForAsync(() => {
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
