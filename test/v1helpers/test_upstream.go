package v1helpers

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	static_plugin_gloo "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/static"
	testgrpcservice "github.com/solo-io/gloo/test/v1helpers/test_grpc_service"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

type ReceivedRequest struct {
	Method      string
	Body        []byte
	Host        string
	GRPCRequest proto.Message
	Port        uint32
}

func NewTestHttpUpstream(ctx context.Context, addr string) *TestUpstream {
	backendPort, responses := runTestServer(ctx, "")
	return newTestUpstream(addr, []uint32{backendPort}, responses)
}

func NewTestHttpUpstreamWithReply(ctx context.Context, addr, reply string) *TestUpstream {
	backendPort, responses := runTestServer(ctx, reply)
	return newTestUpstream(addr, []uint32{backendPort}, responses)
}

func NewTestGRPCUpstream(ctx context.Context, addr string, replicas int) *TestUpstream {
	grpcServices := make([]*testgrpcservice.TestGRPCServer, replicas)
	for i := range grpcServices {
		grpcServices[i] = testgrpcservice.RunServer(ctx)
	}
	received := make(chan *ReceivedRequest, 100)
	for _, srv := range grpcServices {
		srv := srv
		go func() {
			defer GinkgoRecover()
			for r := range srv.C {
				received <- &ReceivedRequest{GRPCRequest: r, Port: srv.Port}
			}
		}()
	}
	ports := make([]uint32, 0, len(grpcServices))
	for _, v := range grpcServices {
		ports = append(ports, v.Port)
	}

	us := newTestUpstream(addr, ports, received)
	us.GrpcServers = grpcServices
	return us
}

type TestUpstream struct {
	Upstream    *gloov1.Upstream
	C           <-chan *ReceivedRequest
	Address     string
	Port        uint32
	GrpcServers []*testgrpcservice.TestGRPCServer
}

func (tu *TestUpstream) FailGrpcHealthCheck() *testgrpcservice.TestGRPCServer {
	for _, v := range tu.GrpcServers[:len(tu.GrpcServers)-1] {
		v.HealthChecker.Fail()
	}
	return tu.GrpcServers[len(tu.GrpcServers)-1]
}

var id = 0

func newTestUpstream(addr string, ports []uint32, responses <-chan *ReceivedRequest) *TestUpstream {
	id += 1
	hosts := make([]*static_plugin_gloo.Host, len(ports))
	for i, port := range ports {
		hosts[i] = &static_plugin_gloo.Host{
			Addr: addr,
			Port: port,
		}
	}
	u := &gloov1.Upstream{
		Metadata: core.Metadata{
			Name:      fmt.Sprintf("local-%d", id),
			Namespace: "default",
		},
		UpstreamSpec: &gloov1.UpstreamSpec{
			UpstreamType: &gloov1.UpstreamSpec_Static{
				Static: &static_plugin_gloo.UpstreamSpec{
					Hosts: hosts,
				},
			},
		},
	}

	return &TestUpstream{
		Upstream: u,
		C:        responses,
		Port:     ports[0],
	}
}

func runTestServer(ctx context.Context, reply string) (uint32, <-chan *ReceivedRequest) {
	bodyChan := make(chan *ReceivedRequest, 100)
	handlerFunc := func(rw http.ResponseWriter, r *http.Request) {
		var rr ReceivedRequest
		rr.Method = r.Method
		if reply != "" {
			_, _ = rw.Write([]byte(reply))
		} else if r.Body != nil {
			body, _ := ioutil.ReadAll(r.Body)
			_ = r.Body.Close()
			if len(body) != 0 {
				rr.Body = body
				_, _ = rw.Write(body)
			}
		}

		rr.Host = r.Host

		bodyChan <- &rr
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	addr := listener.Addr().String()
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handlerFunc))
	mux.Handle("/health", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("OK"))
	}))

	go func() {
		defer GinkgoRecover()
		h := &http.Server{Handler: mux}
		go func() {
			defer GinkgoRecover()
			if err := h.Serve(listener); err != nil {
				if err != http.ErrServerClosed {
					panic(err)
				}
			}
		}()

		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_ = h.Shutdown(ctx)
		cancel()
		// close channel, the http handler may panic but this should be caught by the http code.
		close(bodyChan)
	}()
	return uint32(port), bodyChan
}

func TestUpstreamReachable(envoyPort uint32, tu *TestUpstream, rootca *string) {
	body := []byte("solo.io test")

	ExpectHttpOK(body, rootca, envoyPort, "")

	timeout := time.After(5 * time.Second)
	var receivedRequest *ReceivedRequest
	for {
		select {
		case <-timeout:
			if receivedRequest != nil {
				fmt.Fprintf(GinkgoWriter, "last received request: %v", *receivedRequest)
			}
			Fail("timeout testing upstream reachability")
		case receivedRequest = <-tu.C:
			if receivedRequest.Method == "POST" &&
				bytes.Equal(receivedRequest.Body, body) {
				return
			}
		}
	}

}

func ExpectHttpOK(body []byte, rootca *string, envoyPort uint32, response string) {

	var res *http.Response
	EventuallyWithOffset(2, func() error {
		// send a request with a body
		var buf bytes.Buffer
		buf.Write(body)

		var client http.Client

		scheme := "http"
		if rootca != nil {
			scheme = "https"
			caCertPool := x509.NewCertPool()
			ok := caCertPool.AppendCertsFromPEM([]byte(*rootca))
			if !ok {
				return fmt.Errorf("ca cert is not OK")
			}

			client.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:            caCertPool,
					InsecureSkipVerify: true,
				},
			}
		}

		var err error
		res, err = client.Post(fmt.Sprintf("%s://%s:%d/1", scheme, "localhost", envoyPort), "application/octet-stream", &buf)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("%v is not OK", res.StatusCode)
		}

		return nil
	}, "10s", ".5s").Should(BeNil())

	if response != "" {
		body, err := ioutil.ReadAll(res.Body)
		ExpectWithOffset(2, err).NotTo(HaveOccurred())
		defer res.Body.Close()
		ExpectWithOffset(2, string(body)).To(Equal(response))
	}
}
