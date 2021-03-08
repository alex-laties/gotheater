import { Component } from 'react';
import { Button } from '@material-ui/core';
import { Player, ControlBar } from 'video-react';
import '../node_modules/video-react/dist/video-react.css';
import { TextField } from '@material-ui/core';

export default class VidjaPlayer extends Component {
    constructor(props) {
        super(props);
        this.state = {
            source: 'https://storage.googleapis.com/watchwithme/Contagion.2011.720p.BluRay.x264.AAC-%5BYTS.MX%5D.mp4',
        };

        this.play = this.play.bind(this);
        this.pause = this.pause.bind(this);
    }

    componentDidMount() {
        this.player.subscribeToStateChange(this.handlePlayerStateChange.bind(this));
    }

    componentDidUpdate(prevProps, prevState) {
        if(this.state.source != prevState.source) {
            this.player.load();
        }
    }

    setMuted(muted) {
        this.player.muted = muted;
    }

    handlePlayerStateChange(playerState) {
        this.player = playerState;
    }

    play() {
        this.player.play();
    }

    pause() {
        this.player.pause();
    }

    jumpTo(seconds) {
        this.player.seek(seconds);
    }

    changePlaybackRate(rate) {
        this.player.playbackRate = rate;
    }

    changeVolume(vol) {
        this.player.volume = vol;
    }

    setURL(url) {
        try {
            const url = new URL(url);
        } catch (e) {
            console.error('encountered when loading url', url, e)
            return;
        }
        console.log('switching video from:', this.player.currentSrc, 'to:', url)
        this.setState({
            source: url
        });
    }

    loadField(e) {
        const toLoad = e.target.value;
        this.setURL(toLoad);
    }

    render() {
        return (
            <div>
                <Player
                    ref={player => {this.player = player; }}
                    autoplay
                >
                    <source src={this.state.source} />
                    <ControlBar autoHide={false} />
                </Player>
                <Button onClick={this.play}>Play</Button>
                <Button onClick={this.pause}>Pause</Button>
                <form>
                    <TextField
                        inputRef={urlField => {this.urlField = urlField; }}
                        id="toLoad"
                        label="Standard"
                        value={this.state.source}
                        onChange={this.loadField}
                        />
                </form>
            </div>
        )
    }
}