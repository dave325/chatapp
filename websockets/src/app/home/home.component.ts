import { Component, HostListener, OnDestroy, OnInit } from '@angular/core';
import { map, tap, catchError, retry } from 'rxjs/operators';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';
import { HttpClient, HttpHeaders } from '@angular/common/http';

interface UserMessage {
  user: string;
  message: string;
  room: string;
}

interface Chat {
  key: string;
  chat: WebSocketSubject<any>;
}

interface Room {
  id: string;
  users: string[];
  messages: UserMessage[];
}
interface ResponseData extends Object {
  Total: number;
}

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class HomeComponent implements OnInit {
  title = 'websockets';
  users: Array<string> = [];
  message: UserMessage = { user: "1", message: "", room: "" };
  messages: Map<string, Array<UserMessage>> = new Map<string, Array<UserMessage>>();
  sockets: Map<string, WebSocketSubject<any>> = new Map<string, WebSocketSubject<any>>();
  currentChat: string = "";
  myWebSocket: WebSocketSubject<any>;
  totalUsers: string = "";
  communicaitonSocket: WebSocketSubject<any>;

  constructor(private http: HttpClient) {
  }

  async ngOnInit(): Promise<void> {
    //Called after the constructor, initializing input properties, and the first call to ngOnChanges.
    //Add 'implements OnInit' to the class.
    const existingUser = JSON.parse(window.sessionStorage.getItem('X-DAVE-TEST') || "{}");
    const userExist = window.sessionStorage.getItem('X-DAVE-TEST') ? existingUser : null;
    console.log(existingUser)
    console.log(userExist)
    let user: any;
    if (userExist) {
      user = await this.http.get("http://localhost:3001/users/?user=" + existingUser.username).toPromise();
    } else {
      user = await this.http.get("http://localhost:3001/users/").toPromise();
    }
    console.log(user)

    this.totalUsers = userExist ? existingUser.username : user.currentUser;
    this.users = user.users.filter((user: string) => user.length);

    this.myWebSocket = webSocket({
      url: "ws://localhost:3001/userList/?room=1",
      resultSelector: (data) => {
        console.log(data)
      },
    });
    this.myWebSocket.next({ user: this.totalUsers.toString() });
    this.myWebSocket.subscribe(
      msg => {
        console.log(msg);
        if (!this.users.includes(msg.user.toString())) {
          this.users.push(msg.user.toString());
        }
      },
      err => console.log("dsfs" + err)
    )

    // this.communicaitonSocket = webSocket({
    //   url: "ws://localhost:3001/messages/",
    //   resultSelector:(data) => {
    //     console.log(data)
    //   },
    //   openObserver: {
    //     next: (data) => {
    //       console.log(data)
    //     }
    //   },
    // });

    // this.communicaitonSocket.subscribe(
    //   msg => {
    //     console.log(msg);
    //     if(msg.message.length){
    //       this.messages.push({user: msg.user, message: msg.message});
    //     }
    //   },
    //   err => console.log("dsfs" + err)
    // )

  }

  sendMessage(room: string): void {
    // console.log(this.sockets.get(this.currentChat))
    // this.sockets.get(this.currentChat)?.chat.next(this.message);
    // this.message = {user: this.currentChat, message:""};
    if (this.message.message.length === 0 || !this.sockets.has(room)) {
      console.log("Naaaaa")
      return;
    }
    this.message.room = room;
    this.sockets.get(room)?.next(this.message);
    // tslint:disable-next-line: radix
    this.message = { user: this.totalUsers.toString(), message: "", room: "" };
  }

  async connectUser(user: string): Promise<void> {
    const httpOptions = {
      headers: new HttpHeaders({
        'Content-Type': 'application/json',
        'Accept': "*/*"
      })
    };
    if(user === this.totalUsers){
      return
    }
    const checkChat: Room = await this.http.post<Promise<Room>>(
      "http://localhost:3001/checkChat/",
      { "users": [user, this.totalUsers.toString()] },
      httpOptions).toPromise();
    console.log(checkChat.id)
    if(this.sockets.has(checkChat.id)){
      return
    }

    const currentSocket = await webSocket({
      url: "ws://localhost:3001/messages/?room=" + checkChat.id,
      resultSelector: (data) => {
        console.log(data)
      },
    });
    this.sockets.set(checkChat.id, currentSocket);
    this.messages.set(checkChat.id, checkChat.messages);

    this.sockets.get(checkChat.id)?.subscribe(
      msg => {
        console.log(msg);
        this.messages.get(checkChat.id)?.push({ user: msg.user.username, message: msg.message, room: checkChat.id });
      },
      err => console.log(err)
    )


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
  @HostListener('window:beforeunload')
  async ngOnDestroy(): Promise<void> {
    //Called once, before the instance is destroyed.
    //Add 'implements OnDestroy' to the class.
    for (const [key, chat] of this.sockets) {
      chat.unsubscribe();
      chat.complete();
    }
    const user = await this.http.get("http://localhost:3001/user-unavailable/?user=" + this.totalUsers).toPromise();
    console.log(user)
    await this.myWebSocket.unsubscribe();
    await this.communicaitonSocket.unsubscribe();
  }

}
