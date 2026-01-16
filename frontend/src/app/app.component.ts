import { Component, OnInit } from '@angular/core';
import { RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet],
  template: `
    <router-outlet></router-outlet>
  `
})
export class AppComponent implements OnInit {
  title = 'Spectrum Club Calendar';
  
  ngOnInit() {
    console.log('AppComponent ngOnInit called');
  }
}
