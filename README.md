# music_websocket

A simple websocket server for controlling playing music via UI or API.

## How to run

```bash
git clone github.com/tuanhuu162/music_websocket
cd music_websocket
docker build -t music_websocket .
docker run -d -p 8080:8080 music_websocket
```

Then open your browser and go to `localhost:8080` to see the UI.

## Features
- [x] Play music
- [ ] Pause music
- [ ] Stop music
- [ ] Next music
- [ ] Previous music
- [ ] Play music by name
- [ ] Search music by name
- [ ] Download music