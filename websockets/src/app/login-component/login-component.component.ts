import { HttpClient } from '@angular/common/http';
import { Component, Input, OnInit, Output, EventEmitter } from '@angular/core';
import { FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
interface User {
  id: string;
  password: string;
  token: string;
  username: string;
}
@Component({
  selector: 'app-login-component',
  templateUrl: './login-component.component.html',
  styleUrls: ['./login-component.component.css']
})
export class LoginComponentComponent {
  @Input() error: string | null;

  loginForm: FormGroup;
  submitted = false;

  constructor(private formBuilder: FormBuilder, private router: Router, private http: HttpClient) { }
  form: FormGroup = new FormGroup({
    username: new FormControl('', [Validators.required, Validators.minLength(3)]),
    password: new FormControl('', [Validators.required, Validators.minLength(3)]),
  });

  async submit() {
    if (this.form.valid) {
      // Mark error here 
      const httpCall: User = await this.http.post<User>("http://localhost:3001/login/", this.form.value).toPromise()
      console.log(httpCall)
      window.sessionStorage.setItem("X-DAVE-TEST", JSON.stringify(httpCall))
      this.router.navigate(["/home"])
    }
  }

}
