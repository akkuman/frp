package net

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/akkuman/websocket"
)

var (
	ErrWebsocketListenerClosed = errors.New("websocket listener closed")
)

const (
	FrpWebsocketPath = "/index.js"
)

type WebsocketConfig struct {
	Addr       string
	Origin     string
	SourceHost string
	IsSecure   bool
}

// TODO: support config file
// only for command version, because this var is global
var WSConf WebsocketConfig

type WebsocketListener struct {
	ln       net.Listener
	acceptCh chan net.Conn

	server    *http.Server
	httpMutex *http.ServeMux
}

// NewWebsocketListener to handle websocket connections
// ln: tcp listener for websocket connections
func NewWebsocketListener(ln net.Listener) (wl *WebsocketListener) {
	wl = &WebsocketListener{
		acceptCh: make(chan net.Conn),
	}

	muxer := http.NewServeMux()
	muxer.Handle(FrpWebsocketPath, websocket.Handler(func(c *websocket.Conn) {
		notifyCh := make(chan struct{})
		conn := WrapCloseNotifyConn(c, func() {
			close(notifyCh)
		})
		wl.acceptCh <- conn
		<-notifyCh
	}))

	wl.server = &http.Server{
		Addr:    ln.Addr().String(),
		Handler: muxer,
	}

	go wl.server.Serve(ln)
	return
}

func ListenWebsocket(bindAddr string, bindPort int) (*WebsocketListener, error) {
	tcpLn, err := net.Listen("tcp", fmt.Sprintf("%s:%d", bindAddr, bindPort))
	if err != nil {
		return nil, err
	}
	l := NewWebsocketListener(tcpLn)
	return l, nil
}

func (p *WebsocketListener) Accept() (net.Conn, error) {
	c, ok := <-p.acceptCh
	if !ok {
		return nil, ErrWebsocketListenerClosed
	}
	return c, nil
}

func (p *WebsocketListener) Close() error {
	return p.server.Close()
}

func (p *WebsocketListener) Addr() net.Addr {
	return p.ln.Addr()
}

// addr: domain:port
func ConnectWebsocketServer(addr string) (net.Conn, error) {
	addr = "ws://" + addr + FrpWebsocketPath
	uri, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	origin := "http://" + uri.Host
	cfg, err := websocket.NewConfig(addr, origin)
	if err != nil {
		return nil, err
	}
	cfg.Dialer = &net.Dialer{
		Timeout: 10 * time.Second,
	}

	conn, err := websocket.DialConfig(cfg)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// ConnectWebsocketServerWithCfg Connect Websocket Server with special config
func ConnectWebsocketServerWithCfg(addr string) (net.Conn, error) {
	var err error
	var cfg *websocket.Config

	if WSConf.IsSecure {
		addr = "wss://" + addr + FrpWebsocketPath
	} else {
		addr = "ws://" + addr + FrpWebsocketPath
	}
	uri, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	var origin string
	if WSConf.IsSecure {
		origin = "https://" + uri.Host
	} else {
		origin = "http://" + uri.Host
	}

	options := []websocket.ConfigOption{websocket.WithSourceHost(WSConf.SourceHost)}
	if WSConf.IsSecure {
		options = append(options, websocket.WithTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
		}))
	}

	cfg, err = websocket.NewConfig(addr, origin, options...)
	if err != nil {
		return nil, err
	}

	cfg.Dialer = &net.Dialer{
		Timeout: 10 * time.Second,
	}
	conn, err := websocket.DialConfig(cfg)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// SetWebsocketConfig 设置一些 websocket 基础配置
func SetWebsocketConfig(sourceHost string, isSecure bool) {
	WSConf = WebsocketConfig{
		SourceHost: sourceHost,
		IsSecure:   isSecure,
	}
}
