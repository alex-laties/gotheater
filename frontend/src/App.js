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
import CircularProgress from '@material-ui/core/CircularProgress';
import { Button, Input, TextField } from '@material-ui/core';
import { uniqueNamesGenerator, adjectives, animals } from 'unique-names-generator';

export default class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      id: '',
      rulerID: '',
      paused: false,
      name: uniqueNamesGenerator({dictionaries: [adjectives, animals], length: 2}),
      connected: false,
      messagesReceived: 0,
      messagesSent: 0,
      rtt: 0,
      currUsers: [],
    }
    this.client = null;
    this.keepaliveTimerID = 0;
    this.currentMediaURL = '';
    this.sendMessage = this.sendMessage.bind(this);
    this.setName = this.setName.bind(this);
    this.pushStatus = this.pushStatus.bind(this);
    this.keepAlive = this.keepAlive.bind(this);
    this.setMediaURL = this.setMediaURL.bind(this);
    this.setNewRuler = this.setNewRuler.bind(this);
    this.handlePlayerStateChange = this.handlePlayerStateChange.bind(this);
    this.urlInput = null;
    this.lastSeekTime = 0;
  }

  componentDidMount() {
    let secureWebSockets = false;
    if (window.location.protocol === 'https:') {
      secureWebSockets = true;
    }
    let protocol = secureWebSockets ? 'wss' : 'ws';

    let port = window.location.port ? `:${window.location.port}`: '';

    let wsURL = `${protocol}://${window.location.hostname}${port}/ws`;
    this.initClient(wsURL);
    this.player.player.subscribeToStateChange(this.handlePlayerStateChange);
  }

  handlePlayerStateChange(state, prevState) {
    if (prevState.paused !== state.paused) {
      if (state.paused) {
        this.sendMessage('pause', {});
      } else {
        this.sendMessage('play', {});
      }
      return;
    }
    if (prevState.seeking === false && state.seeking && Math.abs(Date.now() - this.lastSeekTime) > 1000) {
      console.log('seeking', state);
      this.sendMessage('seek', {mediaTimestamp: Math.round(state.currentTime * 1000)});
    }
  }

  componentWillUnmount() {
    if (this.client) {
      this.client.close();
      this.client = null;
    }
    this.cancelKeepAlive();
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
    if (this.client === null) {
      this.keepaliveTimerID = setTimeout(this.keepAlive, 1000)
      return;
    }
    if (this.state.id === '') {
      this.keepaliveTimerID = setTimeout(this.keepAlive, 1000)
      return;
    }
    if (this.client.readyState === this.client.OPEN) {
      this.sendMessage('ping', {timestamp: Date.now()});
      let state = this.player.getState().player;
      this.pushStatus(state);
      if (this.state.id === this.state.rulerID) {
        this.sendMessage('playbackStatus', {
          playing: state.paused,
          currentMediaTimestamp: Math.round(state.currentTime * 1000),
          currentPing: Math.round(this.state.rtt / 2),
        });
      }
    }
    this.keepaliveTimerID = setTimeout(this.keepAlive, 1000)
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

  async handleClientMessage(e) {
    if (this.player === null) {
      return
    }
    this.setState((state) => {
      return {messagesReceived: state.messagesReceived + 1};
    });
    let message = null;
    try {
      let txt;
      if (e.data instanceof Blob) { // sometimes we get a promise of text rather than text...
        txt = await e.data.text();
      } else {
        txt = e.data;
      }
      message = JSON.parse(txt);
    } catch {
      console.log('message plaintext', e.data);
      return;
    }
    switch (message.type) {
    case 'pong':
      let currTime = Date.now();
      let rtt = currTime - message.data.timestamp;
      this.setState({rtt: rtt});
      break;
    case 'connect':
      if (this.state.id === '') {
        this.setState({
          id: message.data.id,
          rulerID: message.data.currentRulerID,
          paused: message.data.currentMediaPaused,
        })
      }
      this.setState({
          currUsers: message.data.currentSessions.filter(el => el.id !== this.state.id)
      });
      if (this.currentMediaURL !== message.data.currentMediaURL) {
        this.currentMediaURL = message.data.currentMediaURL;
        this.player.setURL(this.currentMediaURL);
      }
      if (message.data.currentMediaTimestamp > 0 && Math.abs(Date.now() - this.lastSeekTime) > 1000) {
        this.lastSeekTime = Date.now();
        this.player.jumpTo(message.data.currentMediaTimestamp / 1000.0);
      }
      if (message.data.currentMediaPaused) {
        this.player.pause();
      } else {
        this.player.play();
      }
      break;
    case 'status':
      if (message.id === this.state.id) {
        break;
      }
      const matchingUserIndex = this.state.currUsers.findIndex((el) => el.id === message.id);
      if (matchingUserIndex >= 0) {
        let user = this.state.currUsers[matchingUserIndex];
        user.name = message.data.name;
        user.playing = message.data.playing;
        user.currentMediaURL = message.data.currentMediaURL;
        user.currentMediaTimestamp = message.data.currentMediaTimestamp;
        user.currentPing = message.data.currentPing;
        user.currentPlaybackRate = message.data.currentPlaybackRate;
        this.state.currUsers[matchingUserIndex] = user;
        if (user.id == this.state.rulerID) {
          // adjust playback rate accordingly
          // project the timestamp received
          let timestampShouldBe = user.currentMediaTimestamp + user.currentPing + (this.state.rtt /2);
          let currTimestamp = this.player.getState().player.currentTime * 1000;
          // no need to do anything if we're within 10 milliseconds of projected timestamp
          console.log('drift', timestampShouldBe, currTimestamp, Math.abs(currTimestamp - timestampShouldBe));
          if (Math.abs(currTimestamp - timestampShouldBe) > 10) {
            if (Math.abs(currTimestamp-timestampShouldBe) > 5000 && Math.abs(Date.now() - this.lastSeekTime) > 1000) { // if we're off by 5+ seconds
              this.lastSeekTime = Date.now();
              this.player.jumpTo(timestampShouldBe / 1000); // just jump
            } else if (currTimestamp - timestampShouldBe < 0) { // if we're behind
              console.log("speeding up");
              this.player.changePlaybackRate(1.03); // 3% faster
            } else {
              console.log("slowing down");
              this.player.changePlaybackRate(0.97); // 3% slower
            }
          } else {
            this.player.changePlaybackRate(1);
          }
        }
      }
      break;
    case 'setRuler':
      this.setState({rulerID: message.data.newRulerID});
      break;
    case 'pause':
      console.log(message.type, message.data);
      this.pausing = true;
      this.setState({paused: true});
      this.player.pause();
      break;
    case 'play':
      console.log(message.type, message.data);
      this.setState({paused: true});
      this.player.play();
      break;
    case 'seek':
      console.log(message.type, message.data);
      if (this.applyingSeek) {
        console.log("already applying a seek")
        break;
      }
      this.lastSeekTime = Date.now();
      this.player.jumpTo(message.data.mediaTimestamp / 1000.0);
      break;
    case 'setMedia':
      this.player.setURL(message.data.url);
      break;
    case 'setLeader':
      this.setState({rulerID: message.data.newRulerID});
      break;
    default:
      console.log('notimplemented', message.type);
    }
  }

  pushStatus(vidstate) {
    if (vidstate === null) {
      return;
    }
    let statusData = {
      name: this.state.name,
      playing: !vidstate.paused,
      currentMediaURL: vidstate.currentSrc,
      currentMediaTimestamp: Math.round(vidstate.currentTime * 1000),
      currentPing: Math.round(this.state.rtt / 2),
      currentPlaybackRate: vidstate.playbackRate,
    };

    this.sendMessage('status', statusData);
  }

  handleClientOpen(e) {
    console.log('open', e);
    this.setState({connected: true});
    this.keepAlive();
  }

  setName(e) {
    let name = e.target.value;
    this.setState({name: name});
  }

  setMediaURL(e) {
    if (this.urlInput.value === '') {
      return
    }
    console.log('setMediaURL', this.urlInput);
    const newURL = this.urlInput.value;
    try {
      this.player.setURL(newURL);
    } catch (e) {
      console.log(e);
      return;
    }

    this.currentMediaURL = newURL;
    if (this.state.id === this.state.rulerID) {
      this.sendMessage('setMedia', {url: newURL});
    }
  }

  setNewRuler(id) {
    let toReturn = (e) => {
      this.sendMessage('setLeader', {newRulerID: id});
      this.setState({rulerID: id});
    };
    toReturn.bind(this);
    return toReturn;
  }

  render() {
    let connectedStatus;
    if (!this.state.connected) {
      connectedStatus = <div>connecting <CircularProgress /></div>;
    } else {
      connectedStatus = <div>
        id: {this.state.id} <br />
        name: <TextField value={this.state.name}
                onChange={this.setName}
              /> <br />
        clientStatus: {this.client.readyState} <br />
        messagesReceived: {this.state.messagesReceived} <br />
        messagesSent: {this.state.messagesSent} <br />
        roundTripTime: {this.state.rtt} ms<br />
        amRuler: {this.state.rulerID === this.state.id ? 'yes' : 'no'} <br />
      </div>;
    }
    let rulerDash;
    if (this.state.id === this.state.rulerID) {
      rulerDash = <div>
        <Input
          inputRef={input => { this.urlInput = input; }}
        />
        <Button variant="contained" color="primary" onClick={this.setMediaURL}>Load URL</Button>
      </div>;
    } else {
      rulerDash = <div />
    }
    return (
      <div className="App">
        <VidjaPlayer ref={player => { this.player = player; }} />
        {rulerDash}
        <Box>
          <Card variant="outlined">
            <CardContent>
              {connectedStatus}
            </CardContent>
          </Card>
          {this.state.currUsers.map((val, idx) =>
            <Card key={idx}>
              <CardContent>
                <div>
                  name: {val.name} <br />
                  id: {val.id} <br />
                  status: {val.playing ? "playing" : "paused" } <br />
                  currentMedia: {val.currentMediaURL} <br />
                  currentTimestamp: {Math.floor(val.currentMediaTimestamp/1000/60)}:{Math.floor(val.currentMediaTimestamp/1000%60).toLocaleString('en-US', {minimumIntegerDigits: 2, useGrouping: false})} <br />
                  isRuler: {this.state.rulerID === val.id ? 'yes' : 'no'} <br />
                  currentPing: {val.currentPing} <br />
                  playbackRate: {val.currentPlaybackRate} <br />
                  { this.state.rulerID !== val.id &&
                    <Button variant="contained" color="secondary" onClick={this.setNewRuler(val.id)}>Set as Leader</Button>
                  }
                </div>
              </CardContent>
            </Card>
          )}
        </Box>
      </div>
    );
  }
}