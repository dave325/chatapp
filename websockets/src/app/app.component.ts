import { Message } from '@angular/compiler/src/i18n/i18n_ast';
import { Component } from '@angular/core';
import { map, tap, catchError, retry } from 'rxjs/operators';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';

interface UserMessage {
  user: string;
  message: string;
}
@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'websockets';
  myWebSocket: WebSocketSubject<any> = webSocket({
    url: "ws://localhost:3001/echo",
    resultSelector:(data) => {
      console.log(data)
    },
    openObserver: {
      next: (data) => {
        console.log(data)
      }
    },
  });

  message: UserMessage = {user:"1", message:""};
  messages: Array<any> = [];
  constructor(){
    this.myWebSocket.asObservable().subscribe(
      msg => {
        console.log(msg);
        this.messages.push({user: msg.User, message: msg.Message});
      },
      err => console.log("dsfs" + err)
    )
  }

  sendMessage(): void{
    console.log(this.myWebSocket)
    this.myWebSocket.next(this.message);
    this.message = {user:"1", message:""};
  }
}
