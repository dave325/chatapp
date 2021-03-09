import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import {AppComponent } from "./app.component"
import { LoginComponentComponent } from './login-component/login-component.component';
import { RegistrationComponentComponent } from './registration-component/registration-component.component';
const routes: Routes = [
  {
    path: "",
    component: AppComponent
  },
  {
    path:"register",
    component: RegistrationComponentComponent
  },
  {
    path: "login",
    component: LoginComponentComponent
  }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
