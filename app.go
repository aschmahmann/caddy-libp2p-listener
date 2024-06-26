package caddy_libp2p_listener

import (
	"strconv"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(App{})
	httpcaddyfile.RegisterGlobalOption("libp2p", parseAppConfig)
}

// App is the libp2p Caddy app used to configure libp2p nodes.
type App struct {
	// PrivateKey is the location of the private key as a PEM file
	PrivateKey string `json:"private_key,omitempty" caddy:"namespace=libp2p.key"`

	// AdvertiseAmino indicates if the peer's addresses should be adverised in the Amino DHT
	// This is particularly useful for WebTransport and WebRTC which have their certificate hashes rotate
	// even while their peerID remains the same
	AdvertiseAmino bool `json:"advertise_amino,omitempty" caddy:"namespace=libp2p.advertise_amino"`

	logger *zap.Logger
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "libp2p",
		New: func() caddy.Module { return new(App) },
	}
}

func (t *App) Provision(ctx caddy.Context) error {
	t.logger = ctx.Logger(t)
	return nil
}

func (t *App) Start() error {
	return nil
}

func (t *App) Stop() error {
	return nil
}

func parseAppConfig(d *caddyfile.Dispenser, _ any) (any, error) {
	app := &App{}
	if !d.Next() {
		return app, d.ArgErr()

	}

	for d.NextBlock(0) {
		val := d.Val()

		switch val {
		case "advertise_amino":
			if d.NextArg() {
				v, err := strconv.ParseBool(d.Val())
				if err != nil {
					return nil, d.WrapErr(err)
				}
				app.AdvertiseAmino = v
			} else {
				app.AdvertiseAmino = true
			}
		case "private_key":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.PrivateKey = d.Val()
		default:
			return nil, d.Errf("unrecognized directive: %s", d.Val())
		}
	}
	return httpcaddyfile.App{
		Name:  "libp2p",
		Value: caddyconfig.JSON(app, nil),
	}, nil
}

var (
	_ caddy.App         = (*App)(nil)
	_ caddy.Provisioner = (*App)(nil)
)
