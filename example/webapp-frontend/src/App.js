import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';

class App extends Component {
  _timer = 0;

  state = {
    txtInput: '',
    result: '',
    hasErr: false,
    time: ''
  };

  constructor(props) {
    super(props);
    this.refreshTime();
  }

  render() {
    const { txtInput, result, hasErr, time } = this.state;
    return (
      <div className="App">
        <header className="App-header">
          <img src={logo} className="App-logo" alt="logo" />
          <h1 className="App-title">Welcome to React</h1>
        </header>
        <div className="main-content">
          <div className="say-hello">
            <label>
              Enter name:
              <input
                value={txtInput}
                onChange={e => this.setState({ txtInput: e.target.value })}
              />
            </label>
            <input
              type="button"
              value="Submit"
              onClick={() => this.handleSubmit()}
            />
            <div className={'result ' + (hasErr ? 'error' : '')}>{result}</div>
          </div>
          <div className="time">
            <label>Refresh Time:</label>
            <input
              type="button"
              value="On"
              onClick={() => {
                clearInterval(this._timer);
                this._timer = setInterval(() => this.refreshTime(), 1000);
              }}
            />
            <input
              type="button"
              value="Off"
              onClick={() => clearInterval(this._timer)}
            />

            <div className={'result'}>Current time: {time}</div>
          </div>
        </div>
      </div>
    );
  }

  refreshTime() {
    fetch('/api/time')
      .then(res => {
        if (res.ok) {
          res.text().then(data => this.setState({ time: data }));
        } else {
          console.error('failed to parse text');
        }
      })
      .catch(err => console.error(err.message));
  }

  handleSubmit() {
    fetch(`/api/sayhello?name=${this.state.txtInput}`)
      .then(res => {
        if (res.ok) {
          res
            .text()
            .then(data => this.setState({ result: data, hasErr: false }));
        } else {
          this.setState({ result: 'failed to parse text', hasErr: true });
        }
      })
      .catch(err => this.setState({ result: err.message, hasErr: true }));
  }
}

export default App;
