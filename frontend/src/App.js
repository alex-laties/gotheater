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
      connected: false,
      messagesReceived: 0,
      messagesSent: 0,
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
    let wsURL = 'ws://localhost:10000';
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

  sendMessage(message) {
    if (this.client !== null) {
      this.client.send(message);
      this.setState((state) => {
        return {messagesSent: state.messagesSent + 1};
      });
    }
  }

  keepAlive() {
    if (this.client.readyState == this.client.OPEN) {
      this.sendMessage(JSON.stringify({type: 'ping', initTime: Date.now()}));
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
    try {
      console.log('message', JSON.parse(e.data));
    } catch {
      console.log('message plaintext', e.data);
    }

    //TODO route message based on type
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
        messagesSent: {this.state.messagesSent}
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