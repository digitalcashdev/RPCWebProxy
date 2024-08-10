package rpcproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RPCProxy struct {
	BaseURL         string
	Username        string
	Password        string
	AllowedRequests AllowedRPCRequests
}

type AllowedRPCRequests struct {
	Methods map[string][]map[string][]string `json:"methods"`
}

type RPCRequest struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

func AddCORSHandler(w http.ResponseWriter, r *http.Request) {
	addCORS(w, r)
	w.WriteHeader(http.StatusOK)
}

func addCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("origin")
	if len(origin) == 0 {
		host := r.Host
		if len(host) > 0 {
			origin = "https://" + host
		} else {
			origin = "http://localhost"
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", origin) // Replace with your desired origin
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func (p *RPCProxy) AuthAndProxyHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 100*1024)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var rpcReq RPCRequest
	if err := json.Unmarshal(body, &rpcReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(rpcReq.Params) > 0 {
		if !p.isPublicMethodAndArg(rpcReq.Method, rpcReq.Params[0]) {
			msg := fmt.Sprintf("RPC '%s' '%s' is not a known, public API", rpcReq.Method, rpcReq.Params[0])
			http.Error(w, msg, http.StatusForbidden)
			return
		}
	}

	fmt.Printf("%s => %s\n%s\n\n", p.BaseURL, rpcReq.Method, string(body))
	br := bytes.NewReader(body)
	req, err := http.NewRequest("POST", p.BaseURL, br)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(p.Username, p.Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to proxy request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	addCORS(w, r)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (p *RPCProxy) isPublicMethodAndArg(rpcName string, firstArgAny interface{}) bool {
	paramMaps, ok := p.AllowedRequests.Methods[rpcName]
	if !ok {
		return false
	}

	if len(paramMaps) == 0 {
		return true
	}

	if len(paramMaps) > 1 {
		return false // not implemented
	}

	firstArg, ok := firstArgAny.(string)
	if !ok {
		return false // not implemented
	}

	firstSubcommandValuesMap := paramMaps[0]
	for subcommand, subParamMaps := range firstSubcommandValuesMap {
		if firstArg == subcommand {
			return true
		}
		if len(subParamMaps) == 0 {
			continue
		}
		return false // not implemented
	}

	return false
}

func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
