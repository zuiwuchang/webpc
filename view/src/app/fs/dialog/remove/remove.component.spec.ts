import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';

import { RemoveComponent } from './remove.component';

describe('RemoveComponent', () => {
  let component: RemoveComponent;
  let fixture: ComponentFixture<RemoveComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      declarations: [ RemoveComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(RemoveComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
