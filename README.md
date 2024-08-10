# Dash RPC Proxy (+ Explorer)

Web-friendly, CORS-enabled RPC Proxy to the Digital Cash network. \
(RPC Explorer included)

- `mainnet`: <https://rpc.digitalcash.dev>
- `testnet`: <https://trpc.digitalcash.dev>

<kbd><img width="913" alt="screenshot of RPC Proxy + Explorer performing a 'getaddressbalance' request" src="https://github.com/user-attachments/assets/b2860c24-85db-411c-b2a6-f7001cae49f4"></kbd>

## How to Self-Host Proxy + Explorer

0. Clone and enter the repo

   ```sh
   git clone https://github.com/digitalcashdev/rpcproxy.git
   pushd ./rpcproxy/
   ```

1. Install Go **v1.22+**

   ```sh
   curl https://webi.sh/go | sh
   source ~/.config/envman/PATH.env
   ```

2. Season the Explorer to taste

   ```sh
   ls ./static/
   vi ./static/index.html
   ```

3. Build and Run the Proxy \

   ```sh
   # Option 1: to run locally
   go build -o ./dash-rpcproxy ./cmd/dash-rpcproxy/

   # Option 2: to run on a server or in a container
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go build -o ./dash-rpcproxy-linux-x86_64 ./cmd/dash-rpcproxy/
   ```

   ```sh
   ./dash-rpcproxy --help

   ./dash-rpcproxy --port 8080 --web-root ./my-custom-explorer/ \
        --rpc-hostname 'localhost' --rpc-port 19998 \
        --rpc-username 'user' --rpc-password 'secret'
   ```

## How to Register as a System Daemon

0. Install `serviceman`, `pathman`, and `setcap-netbind`

   ```sh
   curl https://webi.sh/ | sh
   source ~/.config/envman/PATH.env

   webi serviceman setcap-netbind
   ```

1. Place `dash-rpcproxy` in your `PATH`

   ```sh
   mkdir -p ~/bin/
   pathman add ~/bin/
   source ~/.config/envman/PATH.env

   mv ./dash-rpcproxy ~/bin/
   ```

2. Allow binding to privileged ports \
   (optional: non-root install on Linux)

   ```sh
   setcap-netbind 'dash-rpcproxy'
   ```

3. Register the service

   ```sh
   sudo env PATH="$PATH" \
       serviceman add --name "dash-rpcproxy" --system --path="$PATH" -- \
       dash-rpcproxy --port 8080 \
           --rpc-hostname 'localhost' --rpc-port 19998 \
           --rpc-username 'user' --rpc-password 'secret'
   ```

   (note: this also works with ENVs, see [./example.env](/example.env))
