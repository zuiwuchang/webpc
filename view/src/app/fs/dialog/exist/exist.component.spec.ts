import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ExistComponent } from './exist.component';

describe('ExistComponent', () => {
  let component: ExistComponent;
  let fixture: ComponentFixture<ExistComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ExistComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ExistComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
