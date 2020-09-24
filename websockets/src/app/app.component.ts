import { Component } from '@angular/core';
import { map, tap, catchError, retry } from 'rxjs/operators';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'websockets';
  myWebSocket: WebSocketSubject<any> = webSocket("ws://localhost:3001/echo");
  constructor(){
    this.myWebSocket.asObservable().subscribe(
      msg => console.log("message received ") + msg,
      err => console.log(err)
    )
  }

  sendMessage(): void{

    this.myWebSocket.next("Hello");
  }
}
