// import { Injectable } from '@angular/core';
// import { webSocket, WebSocketSubject } from 'rxjs/webSocket';
// import { catchError, tap, switchAll } from 'rxjs/operators';
// import { EMPTY, Subject } from 'rxjs';
// export const WS_ENDPOINT = "ws://localhost:3001/echo"

// @Injectable({
//   providedIn: 'root'
// })
// export class DataService {
//   private socket$: WebSocketSubject<any>;
//   private messagesSubject$ = new Subject();
//   public messages$ = this.messagesSubject$.pipe(switchAll(), catchError(e => { throw e }));

//   public connect(): void {

//     if (!this.socket$ || this.socket$.closed) {
//       this.socket$ = this.getNewWebSocket();
//       const messages = this.socket$.pipe(
//         tap({
//           error: error => console.log(error),
//         }), catchError(_ => EMPTY));
//       this.messagesSubject$.next(messages);
//     }
//   }

//   private getNewWebSocket() {
//     return webSocket(WS_ENDPOINT);
//   }
//   sendMessage(msg: any) {
//     console.log("akjhdfdksj")
//     console.log(this.socket$)
//     this.socket$.next(msg);
//   }
//   close() {
//     this.socket$.complete(); }}
