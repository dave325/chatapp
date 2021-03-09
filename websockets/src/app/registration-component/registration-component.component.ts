import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { MatSnackBar } from '@angular/material/snack-bar';

@Component({
  selector: 'app-registration-component',
  templateUrl: './registration-component.component.html',
  styleUrls: ['./registration-component.component.css']
})
export class RegistrationComponentComponent implements OnInit {
  registerForm: FormGroup;
  
  constructor(private formBuilder: FormBuilder, private router: Router, private _snackBar: MatSnackBar) { }

  ngOnInit() {
    this.registerForm = this.formBuilder.group({
      username: ['', Validators.required],
      password: ['', Validators.required]
    });
  }

  get data() { return this.registerForm.controls; }

  onSubmit() {    
    if (this.registerForm.invalid) {
      return;
    } else {
      localStorage.setItem("username", this.data.username.value);
      localStorage.setItem("password", this.data.password.value);
      this._snackBar.open('Register Successfully', 'Success', {
        duration: 2000,
      });
      this.router.navigate(['/login']);
    }
  }
}
