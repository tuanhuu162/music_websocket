package server

import (
	"encoding/json"
	"log"
	"os"

	"github.com/faiface/beep/mp3"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/websocket"

	// Used when "enableJWT" constant is true:
	"github.com/iris-contrib/middleware/jwt"

	"github.com/tuanhuu162/music_websocket/server/models"
)

// values should match with the client sides as well.
const enableJWT = true
const namespace = "default"
const sampleRate = 44100
const channels = 2

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

// if namespace is empty then simply websocket.Events{...} can be used instead.
var serverEvents = websocket.Namespaces{
	namespace: websocket.Events{
		websocket.OnNamespaceConnected: func(nsConn *websocket.NSConn, msg websocket.Message) error {
			ctx := websocket.GetContext(nsConn.Conn)

			log.Printf("[%s] connected to namespace [%s] with IP [%s]",
				nsConn, msg.Namespace,
				ctx.RemoteAddr())
			return nil
		},
		websocket.OnNamespaceDisconnect: func(nsConn *websocket.NSConn, msg websocket.Message) error {
			log.Printf("[%s] disconnected from namespace [%s]", nsConn, msg.Namespace)
			return nil
		},
		"play": func(nsConn *websocket.NSConn, msg websocket.Message) error {
			log.Printf("[%s] sent: %s", nsConn, string(msg.Events))
			fileName := string("./static/nen_va_hoa.mp3")
			f, err := os.Open(fileName)
			if err != nil {
				log.Fatal(err)
				nsConn.Conn.Server().Broadcast(nsConn, websocket.Message{
					Namespace: msg.Namespace,
					Room:      msg.Room,
					Body:      []byte("Error file not found"),
				})
				return nil
			}
			streamer, format, err := mp3.Decode(f)
			if err != nil {
				log.Fatal(err)
				nsConn.Conn.Server().Broadcast(nsConn, websocket.Message{
					Namespace: msg.Namespace,
					Room:      msg.Room,
					Body:      []byte("Error decoding file"),
				})
				return nil
			}
			defer streamer.Close()

			length_in_seconds := int(streamer.Len() / format.SampleRate / format.NumChans / 2)
			for i := 0; i < length_in_seconds; i++ {
				buffer := make([]byte, format.SampleRate*format.NumChans*2)
				streamer.Stream(buffer)
				data, err := json.Marshal(models.Track{
					Name:     fileName,
					Length:   length_in_seconds,
					Position: i,
					Data:     buffer,
				})
				if err != nil {
					log.Fatal(err)
					nsConn.Conn.Server().Broadcast(nsConn, websocket.Message{
						Namespace: msg.Namespace,
						Room:      msg.Room,
						Body:      []byte("Error encoding json data"),
					})
					return nil
				}
				nsConn.Conn.Server().Broadcast(nsConn, websocket.Message{
					Namespace: msg.Namespace,
					Room:      msg.Room,
					Body:      data,
				})
			}

			return nil
		},
	},
}

func NewApp() {
	app := iris.New()
	websocketServer := websocket.New(
		websocket.DefaultGorillaUpgrader,
		serverEvents)

	j := jwt.New(jwt.Config{
		// Extract by the "token" url,
		// so the client should dial with ws://localhost:8080/echo?token=$token
		Extractor: jwt.FromParameter("token"),

		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte("My Secret"), nil
		},

		// When set, the middleware verifies that tokens are signed
		// with the specific signing algorithm
		// If the signing method is not constant the
		// `Config.ValidationKeyGetter` callback field can be used
		// to implement additional checks
		// Important to avoid security issues described here:
		// https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
		SigningMethod: jwt.SigningMethodHS256,
	})

	idGen := func(ctx iris.Context) string {
		if username := ctx.GetHeader("X-Username"); username != "" {
			return username
		}

		return websocket.DefaultIDGenerator(ctx)
	}

	// serves the endpoint of ws://localhost:8080/echo
	// with optional custom ID generator.
	websocketRoute := app.Get("/echo", websocket.Handler(websocketServer, idGen))

	if enableJWT {
		// Register the jwt middleware (on handshake):
		websocketRoute.Use(j.Serve)
	}

	// serves the browser-based websocket client.
	app.Get("/", func(ctx iris.Context) {
		ctx.ServeFile("./static/index.html")
	})

	app.Get("/download/{song_name}", func(ctx iris.Context) {
		songName := ctx.Params().Get("song_name")
		ctx.WriteString(songName)
	})

	return app
}
