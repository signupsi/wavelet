package api

import (
	_ "expvar"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/gorilla/websocket"
	"github.com/perlin-network/noise/network"
	"github.com/perlin-network/wavelet/log"
	"github.com/perlin-network/wavelet/node"
	"github.com/rs/cors"
)

// service represents a service.
type service struct {
	clients  map[string]*ClientInfo
	registry *registry
	wavelet  *node.Wavelet
	network  *network.Network
	upgrader websocket.Upgrader
}

// init registers routes to the HTTP serve mux.
func (s *service) init(mux *http.ServeMux) {
	mux.Handle("/debug/vars", http.DefaultServeMux)

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.HandleFunc(RouteSessionInit, s.wrap(s.sessionInitHandler))

	mux.HandleFunc(RouteLedgerState, s.wrap(s.ledgerStateHandler))

	mux.HandleFunc(RouteTransactionList, s.wrap(s.listTransactionHandler))
	mux.HandleFunc(RouteTransactionPoll, s.wrap(s.pollTransactionHandler))
	mux.HandleFunc(RouteTransactionSend, s.wrap(s.sendTransactionHandler))
	mux.HandleFunc(RouteTransaction, s.wrap(s.getTransactionHandler))

	mux.HandleFunc(RouteContractSend, s.wrap(s.sendContractHandler))
	mux.HandleFunc(RouteContractGet, s.wrap(s.getContractHandler))
	mux.HandleFunc(RouteContractList, s.wrap(s.listContractsHandler))

	mux.HandleFunc(RouteStatsReset, s.wrap(s.resetStatsHandler))

	mux.HandleFunc(RouteAccountGet, s.wrap(s.getAccountHandler))
	mux.HandleFunc(RouteAccountPoll, s.wrap(s.pollAccountHandler))

	mux.HandleFunc(RouteServerVersion, s.wrap(s.serverVersionHandler))
}

// Run runs the API server with a specified set of options.
func Run(net *network.Network, opts Options) {
	plugin, exists := net.Plugin(node.PluginID)
	if !exists {
		panic("ledger plugin not found")
	}

	registry := newSessionRegistry()

	go func() {
		for range time.Tick(10 * time.Second) {
			registry.Recycle()
		}
	}()

	clients := make(map[string]*ClientInfo)

	for _, client := range opts.Clients {
		clients[client.PublicKey] = client
	}

	mux := http.NewServeMux()

	service := &service{
		clients:  clients,
		registry: newSessionRegistry(),
		wavelet:  plugin.(*node.Wavelet),
		network:  net,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	service.init(mux)

	handler := cors.AllowAll().Handler(mux)

	server := &http.Server{
		Addr:    opts.ListenAddr,
		Handler: handler,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg(" ")
	}
}