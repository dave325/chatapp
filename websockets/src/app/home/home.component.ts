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
  singleSelect: boolean = true;
  multiUserList: Set<string> = new Set()
  usersInChat: Map<string,string[]> = new Map<string,string[]>()
  
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
    this.users = user.users.filter((user: string) => user.length && user !== this.totalUsers);

    this.myWebSocket = webSocket({
      url: "ws://localhost:3001/userList/?room=1",
      resultSelector: (data) => {
        console.log(data)
      },
    });
    this.myWebSocket.next({ user: this.totalUsers.toString(), available: true });
    this.myWebSocket.subscribe(
      msg => {
        console.log(msg);
        if (!msg.available && this.users.includes(msg.user.toString())) {
          this.users = this.users.filter(user => user !== msg.user.toString())
          return
        }
        if (msg.available && !this.users.includes(msg.user.toString()) && msg.user.toString() !== this.totalUsers.toString()) {
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
    this.scrollToBottom(room)
  }

  scrollToBottom(id: string ): void {
    try {
      const el: any = document.getElementById(id)
      console.log(el)
      el.scrollTop = el.scrollHeight;
    } catch(err) { }                 
} 
  async connectUser(user: string): Promise<void> {
    const httpOptions = {
      headers: new HttpHeaders({
        'Content-Type': 'application/json',
        'Accept': "*/*"
      })
    };
    if (user === this.totalUsers) {
      return
    }
    const checkChat: Room = await this.http.post<Promise<Room>>(
      "http://localhost:3001/checkChat/",
      { "users": [user, this.totalUsers.toString()] },
      httpOptions).toPromise();
    console.log(checkChat.id)
    if (this.sockets.has(checkChat.id)) {
      return
    }
    console.log(checkChat)
    this.usersInChat.set(checkChat.id, checkChat.users)

    const currentSocket = await webSocket({
      url: "ws://localhost:3001/messages/?room=" + checkChat.id,
      resultSelector: (data) => {
        console.log(data)
      },
    });
    const messages = checkChat.messages ? checkChat.messages : [];
    this.sockets.set(checkChat.id, currentSocket);
    this.messages.set(checkChat.id, messages);

    this.sockets.get(checkChat.id)?.subscribe(
      msg => {
        console.log(msg);
        this.messages.get(checkChat.id)?.push({ user: msg.user, message: msg.message, room: checkChat.id });
      },
      err => console.log(err)
    )
  }

  async connectMultipleUsers() {

    const httpOptions = {
      headers: new HttpHeaders({
        'Content-Type': 'application/json',
        'Accept': "*/*"
      })
    };
    const users = Array.from(this.multiUserList)
    users.push(this.totalUsers.toString())
    console.log(users)
    const checkChat: Room = await this.http.post<Promise<Room>>(
      "http://localhost:3001/checkChat/",
      { "users": users },
      httpOptions).toPromise();
    console.log(checkChat.id)
    if (this.sockets.has(checkChat.id)) {
      return
    }

    console.log(checkChat)
    const currentSocket = await webSocket({
      url: "ws://localhost:3001/messages/?room=" + checkChat.id,
      resultSelector: (data) => {
        console.log(data)
      },
    });
    const messages = checkChat.messages ? checkChat.messages : [];
    this.usersInChat.set(checkChat.id, checkChat.users)

    this.sockets.set(checkChat.id, currentSocket);
    this.messages.set(checkChat.id, messages);

    this.sockets.get(checkChat.id)?.subscribe(
      msg => {
        console.log(msg);
        this.messages.get(checkChat.id)?.push({ user: msg.user.username, message: msg.message, room: checkChat.id });
      },
      err => console.log(err)
    )
    this.singleSelect = !this.singleSelect

  }

  updateMultiUserList(user: string){
    if(this.multiUserList.has(user)){
      this.multiUserList.delete(user)
    }else{
      this.multiUserList.add(user)
    }
  }

  toggleMultUserSelect(){
    this.singleSelect = !this.singleSelect
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
