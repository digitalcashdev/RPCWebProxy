package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/digitalcashdev/rpcproxy"
	"github.com/digitalcashdev/rpcproxy/internal"
	"github.com/digitalcashdev/rpcproxy/static"
)

var allowedRPCRequests rpcproxy.AllowedRPCRequests

func main() {
	var httpPort int

	defaultHTTPPort := 8080
	httpPortStr := os.Getenv("DASHD_HTTP_PORT")
	if len(httpPortStr) > 0 {
		defaultHTTPPort, _ = strconv.Atoi(httpPortStr)
		if defaultHTTPPort == 0 {
			defaultHTTPPort = 8080
		}
	}

	defaultRPCProtocol := "http"
	rpcProtocol := os.Getenv("DASHD_RPC_PROTOCOL")
	if len(rpcProtocol) == 0 {
		rpcProtocol = defaultRPCProtocol
	}

	rpcHostname := os.Getenv("DASHD_RPC_HOSTNAME")
	if len(rpcHostname) == 0 {
		rpcHostname = "localhost"
	}

	defaultRPCPort := 9998
	rpcPortStr := os.Getenv("DASHD_RPC_PORT")
	if len(rpcPortStr) > 0 {
		defaultRPCPort, _ = strconv.Atoi(rpcPortStr)
		if defaultRPCPort == 0 {
			defaultRPCPort = 8080
		}
	}

	overlayFS := &internal.OverlayFS{}
	proxyURL := fmt.Sprintf("%s://%s:%d", rpcProtocol, rpcHostname, defaultRPCPort)
	username := ""
	password := ""
	flag.StringVar(&proxyURL, "rpc-url", proxyURL, "dashd RPC base URL")
	flag.StringVar(&username, "rpc-username", "", "dashd RPC username")
	flag.StringVar(&password, "rpc-password", "", "dashd RPC password")
	flag.StringVar(&overlayFS.WebRoot, "web-root", "./public/", "serve from the given directory")
	flag.BoolVar(&overlayFS.WebRootOnly, "web-root-only", false, "do not serve the embedded web root")
	flag.IntVar(&httpPort, "port", defaultHTTPPort, "bind and listen for http on this port")
	flag.Parse()

	overlayFS.LocalFS = http.Dir(overlayFS.WebRoot)
	overlayFS.EmbedFS = http.FS(static.FS)

	publicRPCJSONPath := "public-rpcs.json"
	f, err := overlayFS.ForceLocalOrEmbedOpen(publicRPCJSONPath)
	if err != nil {
		log.Fatalf("loading RPC JSON description file '%s' failed: %v", publicRPCJSONPath, err)
	}

	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&allowedRPCRequests); err != nil {
		log.Fatalf("decoding %s failed: %v", publicRPCJSONPath, err)
		return
	}

	if len(username) == 0 {
		username = os.Getenv("DASHD_RPC_USERNAME")
	}
	if len(password) == 0 {
		password = os.Getenv("DASHD_RPC_PASSWORD")
	}

	rpcProxy := &rpcproxy.RPCProxy{
		BaseURL:         proxyURL,
		Username:        username,
		Password:        password,
		AllowedRequests: allowedRPCRequests,
	}
	http.HandleFunc("OPTIONS /", rpcproxy.AddCORSHandler)

	limitedProxier := rpcproxy.RateLimitMiddleware(rpcProxy.AuthAndProxyHandler)
	http.HandleFunc("POST /", limitedProxier)
	//http.HandleFunc("POST /rpc/{rpc}", rpcProxy.proxyRPCRequest)

	fileServer := http.FileServer(overlayFS)
	http.Handle("GET /", fileServer)

	http.HandleFunc("/", rpcproxy.MethodNotAllowedHandler)

	fmt.Printf("Proxying to %s as %s\n", rpcProxy.BaseURL, rpcProxy.Username)
	fmt.Printf("Listening on :%d\n", httpPort)
	addr := fmt.Sprintf(":%d", httpPort)
	log.Fatal(http.ListenAndServe(addr, nil))
}
