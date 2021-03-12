import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { AppComponent } from "./app.component"
import { HomeComponent } from './home/home.component';
import { LoginComponentComponent } from './login-component/login-component.component';
import { RegistrationComponentComponent } from './registration-component/registration-component.component';
const routes: Routes = [
  {
    path: "register",
    component: RegistrationComponentComponent
  },
  {
    path: "login",
    component: LoginComponentComponent
  },
  {
    path: "home",
    component: HomeComponent
  },
  { path: '', 
    redirectTo: '/login', 
    pathMatch: 'full' 
  }, // redirect to `first-component`}
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
