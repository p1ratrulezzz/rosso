package main

import (
   "context"
   "fmt"
   "net"
   "net/http"
   "os"
   "os/signal"
   "sync"
   "time"
)

func main() {
	// Providing certain log configuration before Run() is optional
	// e.g. ConfigLogging(lconf) where lconf is a *LogConfig
	pc := NewProxychannel(
		DefaultHandlerConfig,
		DefaultServerConfig,
		make(map[string]Extension))
	pc.Run()
}

// FailEventType .
// When a request is aborted, the event should be one of the following.
const (
	ConnectFail        = "CONNECT_FAIL"
	AuthFail           = "AUTH_FAIL"
	BeforeRequestFail  = "BEFORE_REQUEST_FAIL"
	BeforeResponseFail = "BEFORE_RESPONSE_FAIL"
	ParentProxyFail    = "PARENT_PROXY_FAIL"

	HTTPDoRequestFail               = "HTTP_DO_REQUEST_FAIL"
	HTTPWriteClientFail             = "HTTP_WRITE_CLIENT_FAIL"
	HTTPSGenerateTLSConfigFail      = "HTTPS_GENERATE_TLS_CONFIG_FAIL"
	HTTPSHijackClientConnFail       = "HTTPS_HIJACK_CLIENT_CONN_FAIL"
	HTTPSWriteEstRespFail           = "HTTPS_WRITE_EST_RESP_FAIL"
	HTTPSTLSClientConnHandshakeFail = "HTTPSTLS_CLIENT_CONN_HANDSHAKE_FAIL"
	HTTPSReadReqFromBufFail         = "HTTPS_READ_REQ_FROM_BUF_FAIL"
	HTTPSDoRequestFail              = "HTTPS_DO_REQUEST_FAIL"
	HTTPSWriteRespFail              = "HTTPS_WRITE_RESP_FAIL"
	TunnelHijackClientConnFail      = "TUNNEL_HIJACK_CLIENT_CONN_FAIL"
	TunnelDialRemoteServerFail      = "TUNNEL_DIAL_REMOTE_SERVER_FAIL"
	TunnelWriteEstRespFail          = "TUNNEL_WRITE_EST_RESP_FAIL"
	TunnelConnectRemoteFail         = "TUNNEL_CONNECT_REMOTE_FAIL"
	TunnelWriteTargetConnFinish     = "TUNNEL_WRITE_TARGET_CONN_FINISH"
	TunnelWriteClientConnFinish     = "TUNNEL_WRITE_CLIENT_CONN_FINISH"

	PoolGetParentProxyFail         = "POOL_GET_PARENT_PROXY_FAIL"
	PoolReadRemoteFail             = "POOL_READ_REMOTE_FAIL"
	PoolWriteClientFail            = "POOL_WRITE_CLIENT_FAIL"
	PoolGetConnPoolFail            = "POOL_GET_CONN_POOL_FAIL"
	PoolNoAvailableParentProxyFail = "POOL_NO_AVAILABLE_PARENT_PROXY_FAIL"
	PoolRoundTripFail              = "POOL_ROUND_TRIP_FAIL"
	PoolParentProxyFail            = "POOL_PARENT_PROXY_FAIL"
	PoolHTTPRegularFinish          = "POOL_HTTP_REGULAR_FINISH"
	PoolGetConnFail                = "POOL_GET_CONN_FAIL"
	PoolWriteTargetConnFail        = "POOL_WRITE_TARGET_CONN_FAIL"
	PoolReadTargetFail             = "POOL_READ_TARGET_FAIL"

	HTTPWebsocketDailFail                    = "HTTP_WEBSOCKET_DAIL_FAIL"
	HTTPWebsocketHijackFail                  = "HTTP_WEBSOCKET_HIJACK_FAIL"
	HTTPWebsocketHandshakeFail               = "HTTP_WEBSOCKET_HANDSHAKE_FAIL"
	HTTPSWebsocketGenerateTLSConfigFail      = "HTTPS_WEBSOCKET_GENERATE_TLS_CONFIG_FAIL"
	HTTPSWebsocketHijackFail                 = "HTTPS_WEBSOCKET_HIJACK_FAIL"
	HTTPSWebsocketWriteEstRespFail           = "HTTPS_WEBSOCKET_WRITE_EST_RESP_FAIL"
	HTTPSWebsocketTLSClientConnHandshakeFail = "HTTPS_WEBSOCKET_TLS_CLIENT_CONN_HANDSHAKE_FAIL"
	HTTPSWebsocketReadReqFromBufFail         = "HTTPS_WEBSOCKET_READ_REQ_FROM_BUF_FAIL"
	HTTPSWebsocketDailFail                   = "HTTPS_WEBSOCKET_DAIL_FAIL"
	HTTPSWebsocketHandshakeFail              = "HTTPS_WEBSOCKET_HANDSHAKE_FAIL"

	HTTPRedialCancelTimeout   = "HTTP_REDIAL_CANCEL_TIMEOUT"
	HTTPSRedialCancelTimeout  = "HTTPS_REDIAL_CANCEL_TIMEOUT"
	TunnelRedialCancelTimeout = "TUNNEL_REDIAL_CANCEL_TIMEOUT"
)

// Proxychannel is a prxoy server that manages data transmission between http
// clients and destination servers. With the "Extensions" provided by user,
// Proxychannel is able to do authentication, communicate with databases,
// manipulate the requests/responses, etc.
type Proxychannel struct {
	extensionManager *ExtensionManager
	server           *http.Server
	waitGroup        *sync.WaitGroup
	serverDone       chan bool
}

func NewProxychannel(hconf *HandlerConfig, sconf *ServerConfig, m map[string]Extension) *Proxychannel {
	pc := &Proxychannel{
		extensionManager: NewExtensionManager(m),
		waitGroup:        &sync.WaitGroup{},
		serverDone:       make(chan bool),
	}
	pc.server = NewServer(hconf, sconf, pc.extensionManager)
	return pc
}

func NewServer(hconf *HandlerConfig, sconf *ServerConfig, em *ExtensionManager) *http.Server {
	handler := NewProxy(hconf, em)
	server := &http.Server{
		Addr:         sconf.ProxyAddr,
		Handler:      handler,
		ReadTimeout:  sconf.ReadTimeout,
		WriteTimeout: sconf.WriteTimeout,
		TLSConfig:    sconf.TLSConfig,
	}
	return server
}

func (pc *Proxychannel) runExtensionManager() {
	defer pc.waitGroup.Done()
	go pc.extensionManager.Setup() // TODO: modify setup and error handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan)
	// Will block until shutdown signal is received
	<-signalChan
	// Will block until pc.server has been shut down
	<-pc.serverDone
	pc.extensionManager.Cleanup()
}

func (pc *Proxychannel) runServer() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer close(pc.serverDone)
	pc.server.BaseContext = func(_ net.Listener) context.Context { return ctx }
	stop := func() {
		gracefulCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := pc.server.Shutdown(gracefulCtx); err != nil {
			fmt.Printf("HTTP server Shutdown error: %v\n", err)
		} else {
			fmt.Println("HTTP server gracefully stopped")
		}
	}
	// Run server
	go func() {
		if err := pc.server.ListenAndServe(); err != http.ErrServerClosed {
			//Logger.Errorf("HTTP server ListenAndServe: %v", err)
			os.Exit(1)
		}
	}()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan)
	// Will block until shutdown signal is received
	<-signalChan
	// Terminate after second signal before callback is done
	go func() {
		<-signalChan
		os.Exit(1)
	}()
	stop()
}

// Run launches the ExtensionManager and the HTTP server
func (pc *Proxychannel) Run() {
	pc.waitGroup.Add(1)
	go pc.runExtensionManager()
	pc.runServer()
	pc.waitGroup.Wait()
}
