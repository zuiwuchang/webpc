import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';

import { CompressComponent } from './compress.component';

describe('CompressComponent', () => {
  let component: CompressComponent;
  let fixture: ComponentFixture<CompressComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [ CompressComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(CompressComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
