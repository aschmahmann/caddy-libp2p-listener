package caddy_libp2p_listener

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	libp2phttp "github.com/libp2p/go-libp2p/p2p/http"
	"github.com/libp2p/go-libp2p/p2p/net/gostream"
	libp2pwebrtc "github.com/libp2p/go-libp2p/p2p/transport/webrtc"
	"github.com/multiformats/go-multiaddr"
	"net"
	"os"
	"strconv"
	"strings"
)

func init() {
	//caddy.RegisterModule(new(LL))
	caddy.RegisterNetwork("multiaddr", registerMultiaddrURI)
}

func registerMultiaddrURI(ctx context.Context, network, addr string, cfg net.ListenConfig) (any, error) {
	cctx, ok := ctx.(caddy.Context)
	if !ok {
		return nil, fmt.Errorf("context is not a caddy.Context: %T", ctx)
	}

	if network != "multiaddr" {
		return nil, fmt.Errorf("multiaddr URI network only handles multiaddr URIs")
	}

	lastColon := strings.LastIndex(addr, ":")
	var port int
	var err error
	if lastColon != len(addr)-1 {
		port, err = strconv.Atoi(addr[lastColon+1:])
		if err != nil {
			return nil, fmt.Errorf("invalid port %w", err)
		}
	}
	addr = addr[:lastColon]

	ma, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return nil, fmt.Errorf("could not parse multiaddr: %w", err)
	}

	// TODO: check port matches first one in multiaddr
	_ = port

	/*
		ai, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			return nil, fmt.Errorf("only libp2p multiaddrs currently supported: %w", err)
		}

		if len(ai.Addrs) == 0 {
			return nil, fmt.Errorf("must listen on a supported address type")
		}
	*/

	appIface, err := cctx.App("libp2p")
	if err != nil {
		return nil, err
	}
	app := appIface.(*App)
	var sk crypto.PrivKey
	if app.PrivateKey != "" {
		privFile, err := os.ReadFile(app.PrivateKey)
		if err != nil {
			return nil, err
		}

		pemBlock, rest := pem.Decode(privFile)
		if pemBlock == nil {
			return nil, fmt.Errorf("PEM block not found in input data:\n%s", rest)
		}

		if pemBlock.Type != "PRIVATE KEY" {
			return nil, fmt.Errorf("expected PRIVATE KEY type in PEM block but got: %s", pemBlock.Type)
		}

		stdKey, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parsing PKCS8 format: %w", err)
		}

		// In case ed25519.PrivateKey is returned we need the pointer for
		// conversion to libp2p keys
		if ed25519KeyPointer, ok := stdKey.(ed25519.PrivateKey); ok {
			stdKey = &ed25519KeyPointer
		}

		sk, _, err = crypto.KeyPairFromStdKey(stdKey)
		if err != nil {
			return nil, fmt.Errorf("converting std Go key to libp2p key: %w", err)
		}
	}

	opts := []libp2p.Option{libp2p.ListenAddrs(ma), libp2p.DefaultTransports, libp2p.Transport(libp2pwebrtc.New)}
	if sk != nil {
		opts = append(opts, libp2p.Identity(sk))
	}
	if app.AdvertiseAmino {
		opts = append(opts, libp2p.Routing(func(host host.Host) (routing.PeerRouting, error) {
			r, err := dht.New(ctx, host, dht.Mode(dht.ModeClient))
			return r, err
		}))
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	fmt.Println(h.ID())
	fmt.Println(h.Addrs())

	return gostream.Listen(h, libp2phttp.ProtocolIDForMultistreamSelect)
}

/*
type LL struct {
}

func (L LL) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "caddy.listeners.libp2p",
		New: func() caddy.Module {
			return new(LL)
		},
	}
}
*/
