import { Message } from '@angular/compiler/src/i18n/i18n_ast';
import { Component, OnDestroy, OnInit } from '@angular/core';
import { map, tap, catchError, retry } from 'rxjs/operators';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';
import { HttpClient } from '@angular/common/http';

interface UserMessage {
  user: string;
  message: string;
}

interface Chat {
  key: string;
  chat: WebSocketSubject<any>;
}

interface ResponseData extends Object{
  Total: number;
}

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnDestroy{
  title = 'websockets';
  users: Array<number> = [];
  message: UserMessage = {user:"1", message:""};
  messages: Array<any> = [];
  sockets: Map<string, Chat> = new Map<string, Chat>();
  currentChat: string = "";
  myWebSocket: WebSocketSubject<any>;
  totalUsers: number = -1;
  communicaitonSocket: WebSocketSubject<any>;

  constructor(private http: HttpClient ){
  }

  async ngOnInit(): Promise<void> {
    //Called after the constructor, initializing input properties, and the first call to ngOnChanges.
    //Add 'implements OnInit' to the class.
    const existingUser = JSON.parse(window.sessionStorage.getItem('X-DAVE-TEST') || "{}");
    if(Object.values(existingUser).length === 0){
      const user: any = await this.http.get("http://localhost:3001/users/").toPromise();
      this.totalUsers = user.Total;
    }else{
      console.log(existingUser)
      this.totalUsers = Object.values(existingUser).length > 0 ? parseInt(existingUser) : -1;
      console.log(this.totalUsers)
    }
    this.message.user = this.totalUsers.toString();
    this.myWebSocket = webSocket({
      url: "ws://localhost:3001/userList/?user="+this.totalUsers +"&room=1",
      resultSelector:(data) => {
        console.log(data)
      },
      openObserver: {
        next: (data) => {
          console.log(data)
          window.sessionStorage.setItem('X-DAVE-TEST', this.totalUsers.toString())
          if(!this.users.includes(this.totalUsers)) {
            this.users.push(this.totalUsers);
          }
        }
      },
    });

    this.myWebSocket.subscribe(
      msg => {
        console.log(msg);
        if(msg.Message == undefined){
          this.users.push(msg.User)
          return;
        }
        this.messages.push({user: msg.User, message: msg.Message});
      },
      err => console.log("dsfs" + err)
    )

    this.communicaitonSocket = webSocket({
      url: "ws://localhost:3001/messages/",
      resultSelector:(data) => {
        console.log(data)
      },
      openObserver: {
        next: (data) => {
          console.log(data)
        }
      },
    });

    this.communicaitonSocket.asObservable().subscribe(
      msg => {
        console.log(msg);
        this.messages.push({user: msg.User, message: msg.Message});
      },
      err => console.log("dsfs" + err)
    )

  }
  sendMessage(): void{
    // console.log(this.sockets.get(this.currentChat))
    // this.sockets.get(this.currentChat)?.chat.next(this.message);
    // this.message = {user: this.currentChat, message:""};
    this.communicaitonSocket.next(this.message)
    // tslint:disable-next-line: radix
    this.message = {user: this.totalUsers.toString(), message:""};
  }

  async connectUser(user: string): Promise<void> {

      // for(const userId of this.users){
      //   user = userId.toString();
      //   const socket: WebSocketSubject<any> = await webSocket({
      //       url: 'ws://localhost:3001/messages/?user=' + user,
      //       resultSelector:(data) => {
      //         console.log(data)
      //       },
      //       openObserver: {
      //         next: (data) => {
      //           console.log(data)
      //         }
      //       },
      //     });
      //   const newSocket: Chat = {key: user, chat: socket};
      //   this.sockets.set(user, newSocket);
      //   this.sockets.get(user)?.chat.subscribe(
      //     msg => {
      //       console.log(msg);
      //       this.messages.push({user: msg.User, message: msg.Message});
      //     },
      //     err => console.log("dsfs" + err)
      //   )
      //   this.currentChat = user;
      // }
  }

  ngOnDestroy(): void {
    //Called once, before the instance is destroyed.
    //Add 'implements OnDestroy' to the class.
    for(const [key, chat] of this.sockets){
      chat.chat.unsubscribe();
      chat.chat.complete();
    }
    this.myWebSocket.unsubscribe()
    this.communicaitonSocket.unsubscribe()
  }

}
