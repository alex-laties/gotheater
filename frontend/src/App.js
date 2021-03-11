import logo from './logo.svg';
import './App.css';
import '../node_modules/video-react/dist/video-react.css';
import { Component } from 'react';
import VidjaPlayer from './VidjaPlayer';
import { w3cwebsocket as W3CWebSocket } from "websocket";
import Box from '@material-ui/core/Box';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import Typography from '@material-ui/core/Typography';
import CircularProgress from '@material-ui/core/CircularProgress'

export default class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      id: '',
      rulerID: '',
      currentMediaURL: '',
      currentDrift: 0,
      name: '',
      connected: false,
      messagesReceived: 0,
      messagesSent: 0,
      rtt: 0,
      currUsers: [],
    }
    this.client = null;
    this.keepaliveTimerID = 0;
  }

  componentDidMount() {
    //TODO connect to server and subscribe to events
    let secureWebSockets = false;
    if (window.location.protocol === 'https:') {
      secureWebSockets = true;
    }
    let protocol = secureWebSockets ? 'wss' : 'ws';

    let port = window.location.port ? `:${window.location.port}`: '';

    //let wsURL = `${protocol}://${window.location.hostname}${port}`;
    let wsURL = 'ws://localhost:8080/ws';
    this.initClient(wsURL);
  }

  initClient(url) {
    let client = new W3CWebSocket(url);
    client.onopen = this.handleClientOpen.bind(this);
    client.onerror = this.handleClientError.bind(this);
    client.onmessage = this.handleClientMessage.bind(this);
    client.onclose = this.handleClientError.bind(this);
    this.client = client;
  }

  sendMessage(type, data) {
    if (this.client !== null) {
      let toSend = {
        id: this.state.id,
        type: type,
        data: data
      }

      this.client.send(JSON.stringify(toSend));
      this.setState((state) => {
        return {messagesSent: state.messagesSent + 1};
      });
    }
  }

  keepAlive() {
    if (this.client.readyState == this.client.OPEN) {
      this.sendMessage('ping', {timestamp: Date.now()});
    }
    let kA = this.keepAlive.bind(this)
    this.keepaliveTimerID = setTimeout(kA, 1500)
  }

  cancelKeepAlive() {
    if (this.keepaliveTimerID) {
      clearTimeout(this.keepaliveTimerID)
      this.keepaliveTimerID = 0;
    }
  }

  handleClientClose(e) {
    console.log('close', e);
    this.setState({connected: false});
    this.cancelKeepAlive();
  }

  handleClientError(e) {
    console.log('error', e);
  }

  handleClientMessage(e) {
    this.setState((state) => {
      return {messagesReceived: state.messagesReceived + 1};
    });
    let message = null;
    try {
      message = JSON.parse(e.data);
      console.log('message', message);
    } catch {
      console.log('message plaintext', e.data);
    }
    switch (message.type) {
      case 'pong':
        let currTime = Date.now();
        let rtt = currTime - message.data.timestamp;
        this.setState({rtt: rtt});
        break;
    }
  }

  handleClientOpen(e) {
    console.log('open', e);
    this.setState({connected: true});
    this.keepAlive();
  }

  render() {
    let connectedStatus;
    if (!this.state.connected) {
      connectedStatus = <div>connecting <CircularProgress /></div>;
    } else {
      connectedStatus = <div>
        clientStatus: {this.client.readyState} <br />
        messagesReceived: {this.state.messagesReceived} <br />
        messagesSent: {this.state.messagesSent} <br />
        roundTripTime: {this.state.rtt} ms<br />
      </div>;
    }
    return (
      <div className="App">
        <VidjaPlayer ref={player => { this.player = player; }} />
        <Box>
          <Card variant="outlined">
            <CardContent>
              {connectedStatus}
            </CardContent>
          </Card>
        </Box>
      </div>
    );
  }
}