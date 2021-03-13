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
    }

    getState() {
        return this.player.getState();
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
            let throwaway = new URL(url);
        } catch (e) {
            console.error('encountered when loading url', url, e)
            throw e;
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
                    src={this.state.source}
                    poster="https://external-preview.redd.it/h_toqTwoOJ4LeP1Z2VGXaCO3HujYejJc7uKzZdbPRUA.jpg?auto=webp&s=82b4a93f58ae2770d8ef72d2418b9c34d1835818"
                >
                    <ControlBar autoHide={false} />
                </Player>
            </div>
        )
    }
}