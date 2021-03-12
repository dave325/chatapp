import { HttpClient } from '@angular/common/http';
import { Component, Input } from '@angular/core';
import { FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';

interface User {
  id: string;
  password: string;
  token: string;
  username: string;
}

@Component({
  selector: 'app-registration-component',
  templateUrl: './registration-component.component.html',
  styleUrls: ['./registration-component.component.css']
})
export class RegistrationComponentComponent {
  @Input() error: string | null;

  constructor(private formBuilder: FormBuilder, private router: Router, private http: HttpClient) { }
  form: FormGroup = new FormGroup({
    username: new FormControl('', [Validators.required, Validators.minLength(3)]),
    password: new FormControl('', [Validators.required, Validators.minLength(3)]),
    passwordCheck: new FormControl('', [Validators.required, Validators.minLength(3)]),
  });

  async submit() {
    if (this.form.valid) {
      console.log("yuuurrrr", this.form.value);
      const formValues = this.form.value;
      if (formValues.password === formValues.passwordCheck) {
        // Mark error here 
        delete formValues.passwordCheck;
        const httpCall: User = await this.http.post<User>("http://localhost:3001/sign-up/", this.form.value).toPromise()
        console.log(httpCall)
        const token = httpCall.token
        window.sessionStorage.setItem("X-DAVE-TEST", JSON.stringify(httpCall))
        this.router.navigate(["/home"])
      }
    } else {
      this.error = "Invalid Form"
    }
  }
}
