package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/meyskens/sniproxy/pkg/endpoints"
	"github.com/meyskens/sniproxy/pkg/httpproxy"

	"github.com/meyskens/sniproxy/pkg/sniproxy"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewHostCmd())
}

type hostCmdOptions struct {
	BindAddr string
	Port     int

	sniProxy *sniproxy.SNIProxy
	db       *endpoints.EndpointDB

	httpPort  int
	httpsPort int

	endpointFile string
}

// NewHostCmd generates the `host` command
func NewHostCmd() *cobra.Command {
	s := hostCmdOptions{}
	c := &cobra.Command{
		Use:     "host",
		Short:   "hosts the public endpoint",
		Long:    `hosts the public endpoint on the given bind address`,
		PreRunE: s.Validate,
		RunE:    s.RunE,
	}
	c.Flags().StringVarP(&s.BindAddr, "bind-address", "b", "0.0.0.0", "address to bind port to")
	c.Flags().IntVarP(&s.httpPort, "http-port", "p", 80, "http port to listen on")
	c.Flags().IntVarP(&s.httpsPort, "https-port", "s", 443, "https port to listen on")
	c.Flags().StringVarP(&s.endpointFile, "endpoints-file", "e", "endpoints.txt", "endpoints file")
	return c
}

func (h *hostCmdOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (h *hostCmdOptions) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h.db = endpoints.NewEndpointsDB(ctx, h.endpointFile)
	h.sniProxy = sniproxy.NewSNIProxy(h.db)

	go h.serveTLS()
	go h.serveHTTP()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
			return nil
		}
	}
}

func (h *hostCmdOptions) serveTLS() {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", h.BindAddr, h.httpsPort))
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Started TLS proxy on port", h.httpsPort)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go func() {
			err := h.sniProxy.HandleConnection(conn)
			if err != nil {
				log.Print(err)
			}
		}()
	}
}

func (h *hostCmdOptions) serveHTTP() {
	log.Print("Started HTTP proxy on port", h.httpPort)

	// start server
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		ep, err := h.db.Get(strings.Split(req.Host, ":")[0]) // no port for hostname
		if err != nil || ep == "" {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("no endpoint found"))
			return
		}

		dialer, err := net.Dial("tcp", fmt.Sprintf("%s:80", ep))
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("error dialing endpoint"))
			return
		}
		defer dialer.Close()

		target, err := url.Parse(fmt.Sprintf("http://%s", req.Host))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("error constructing host"))
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.Transport = httpproxy.NewHTTPProxy(dialer)

		proxy.ServeHTTP(w, req)
	})

	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}
