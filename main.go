package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/spf13/cobra"
)

type blockingProxy struct {
	listenPort     int
	targetHostPort string
	targetIP       string
}

var (
	proxy  = blockingProxy{}
	debugLogger = log.New(os.Stdout, "[debug]", log.LstdFlags)
	infoLogger = log.New(os.Stdout, "[info]", log.LstdFlags)
	warnLogger = log.New(os.Stdout, "[warn]", log.LstdFlags)
	errorLogger = log.New(os.Stdout, "[error]", log.LstdFlags)
)

// start listening to the listenPort
func (p *blockingProxy) start() {
	infoLogger.Printf("Starting proxy on port %d", p.listenPort)
	http.ListenAndServe(":"+strconv.Itoa(p.listenPort), p)
}

func httpError(w http.ResponseWriter, message string, code int) {
	errorLogger.Printf("Error: %s", message)
	http.Error(w, message, code)
}

func compareHostPorts(hostPort1 string, hostPort2 string) error {
	host1, port1, err := net.SplitHostPort(hostPort1)
	if err != nil {
		return err
	}

	host2, port2, err := net.SplitHostPort(hostPort2)
	if err != nil {
		return err
	}
	if host1 != host2 || port1 != port2 {
		return fmt.Errorf("Host and port don't match: %s != %s", hostPort1, hostPort2)
	}
	return nil
}

func (p *blockingProxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodConnect {
		httpError(wr, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	debugLogger.Printf("Received request to connect to %s", req.URL.Host) 
	if err := compareHostPorts(req.URL.Host, p.targetHostPort); err != nil {
		httpError(wr, err.Error(), http.StatusBadRequest)
		return
	}
	host := req.URL.Host
	if p.targetIP != "" {
		host = p.targetIP
	}
	debugLogger.Printf("Proxying request to %v", host)
	upstreamConn, err := net.Dial("tcp", host)
	if err != nil {
		httpError(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	wr.WriteHeader(http.StatusOK)
	hijacker, ok := wr.(http.Hijacker)
	if !ok {
		httpError(wr, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	hijackedCon, _, err := hijacker.Hijack()
	if err != nil {
		httpError(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	Pipe(upstreamConn, hijackedCon)
}

func Pipe(conn1 net.Conn, conn2 net.Conn) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer conn1.Close()
		defer conn2.Close()
		if _, err := io.Copy(conn1, conn2); err != nil {
			warnLogger.Printf("error copying data %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		defer conn1.Close()
		defer conn2.Close()
		if _, err := io.Copy(conn2, conn1); err != nil {
			warnLogger.Printf("error copying data %v", err)
		}
	}()

	wg.Wait()

}

var cmd = &cobra.Command{
	Use:   "proxy",
	Short: "A proxy that blocks all requests except for the target host",
	Long:  "A proxy that blocks all requests except for the target host",
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _, err := net.SplitHostPort(proxy.targetHostPort)
		if err != nil {
			warnLogger.Printf("error parsing target host:port %s %s", proxy.targetHostPort, host)
			proxy.targetHostPort = fmt.Sprintf("%s:443", proxy.targetHostPort)
		}
		proxy.start()
		return nil
	},
}

func init() {
	cmd.Flags().IntVarP(&proxy.listenPort, "port", "l", 8080, "The port to listen on")
	cmd.Flags().StringVarP(&proxy.targetHostPort, "target", "t", "", "The host to proxy to (host:port)")
	cmd.MarkFlagRequired("target")
	cmd.Flags().StringVarP(&proxy.targetIP, "target-ip", "i", "", "The ip of the host to proxy to (optional)")
}

func main() {
	cmd.Execute()
}
