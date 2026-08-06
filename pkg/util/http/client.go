package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/paper-indonesia/pdk/v2/logger"
)

const maxResponseSize = 10 << 20 // 10 MB

type HTTPClient interface {
	Request(ctx context.Context, method, url string, body any, headers map[string]string) (resp []byte, code int, err error)

	CloseIdleConnections()
}

type httpClient struct {
	client *http.Client
	pool   *sync.Pool
	logger logger.ILogger
}

type HTTPClientConfig struct {
	DialTimeout         time.Duration `json:"dialTimeout"`
	RequestTimeout      time.Duration `json:"requestTimeout"`
	MaxIdleConns        int           `json:"maxIdleConns"`
	MaxIdleConnsPerHost int           `json:"maxIdleConnsPerHost"`
	IdleConnTimeout     time.Duration `json:"idleConnTimeout"`
	ProbeInterval       time.Duration `json:"probeInterval"`
	MaxProbes           int           `json:"maxProbes"`
}

type HTTPClientOption func(*httpClient)

func NewHTTPClient(config HTTPClientConfig, options ...HTTPClientOption) HTTPClient {
	httpClient := &httpClient{
		client: &http.Client{
			Transport: defaultTransport(config),
		},
		pool: &sync.Pool{
			New: func() any { return new(bytes.Buffer) },
		},
	}
	if config.RequestTimeout > 0 {
		httpClient.client.Timeout = config.RequestTimeout
	}

	for _, option := range options {
		option(httpClient)
	}

	return httpClient
}

func WithTransport(transport *http.Transport) HTTPClientOption {
	return func(hc *httpClient) { hc.client.Transport = transport }
}

func WithDialer(dialer *net.Dialer) HTTPClientOption {
	return func(hc *httpClient) {
		if transport, ok := hc.client.Transport.(*http.Transport); ok {
			transport.DialContext = dialer.DialContext
		}
	}
}

func WithLogger(logger logger.ILogger) HTTPClientOption {
	return func(hc *httpClient) { hc.logger = logger }
}

func (h *httpClient) Request(ctx context.Context, method, uri string, body any, headers map[string]string) ([]byte, int, error) {
	buf := h.pool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		h.pool.Put(buf)
	}()

	if body != nil {
		switch val := body.(type) {
		case []byte:
			if _, err := buf.Write(val); err != nil {
				return nil, 0, err
			}
		case string:
			if _, err := buf.WriteString(val); err != nil {
				return nil, 0, err
			}
		default:
			if err := json.NewEncoder(buf).Encode(val); err != nil {
				return nil, 0, err
			}
		}
	}

	request, err := http.NewRequestWithContext(ctx, method, uri, buf)
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	if val := request.Header.Get(constant.HeaderContentType); val == "" {
		request.Header.Set(constant.HeaderContentType, "application/json")
	}

	request, connInfo := setupHTTPClientTrace(request)

	response, err := h.client.Do(request)
	if err != nil {
		if h.logger == nil {
			log.Printf(
				"ERROR: Failed to perform HTTP request. hostPort=%s reused=%t wasIdle=%t idleTime=%v\n",
				connInfo.HostPort, connInfo.GotConnInfo.Reused, connInfo.GotConnInfo.WasIdle, connInfo.GotConnInfo.IdleTime,
			)
		} else {
			h.logger.Warn(
				ctx, "Failed to perform HTTP request",
				logger.String("hostPort", connInfo.HostPort),
				logger.Bool("reused", connInfo.GotConnInfo.Reused),
				logger.Bool("wasIdle", connInfo.GotConnInfo.WasIdle),
				logger.Duration("idleTime", connInfo.GotConnInfo.IdleTime),
			)
		}
		return nil, 0, err
	}
	defer func() { _ = response.Body.Close() }()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize))
	if err != nil {
		return nil, 0, err
	}
	return responseBody, response.StatusCode, nil
}

func (h *httpClient) CloseIdleConnections() { h.client.CloseIdleConnections() }

type clientConnInfo struct {
	HostPort    string
	GotConnInfo httptrace.GotConnInfo
}

func setupHTTPClientTrace(req *http.Request) (*http.Request, *clientConnInfo) {
	info := &clientConnInfo{}

	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			info.HostPort = hostPort
		},
		GotConn: func(gci httptrace.GotConnInfo) {
			info.GotConnInfo = gci
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	return req, info
}

func defaultTransport(config HTTPClientConfig) *http.Transport {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   http.DefaultMaxIdleConnsPerHost,
	}
	if config.MaxIdleConns > 0 {
		transport.MaxIdleConns = config.MaxIdleConns
	}
	if config.MaxIdleConnsPerHost > 0 {
		transport.MaxIdleConnsPerHost = config.MaxIdleConnsPerHost
	}
	if config.IdleConnTimeout > 0 {
		transport.IdleConnTimeout = config.IdleConnTimeout
	}

	dialTimeout := config.DialTimeout
	if dialTimeout == 0 {
		dialTimeout = 30 * time.Second
	}

	probeInterval := config.ProbeInterval
	if probeInterval == 0 {
		probeInterval = 10 * time.Second
	}

	maxProbes := config.MaxProbes
	if maxProbes == 0 {
		maxProbes = 3
	}

	dialer := &net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: probeInterval,
	}

	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := dialer.DialContext(ctx, network, addr)
		if err != nil {
			log.Printf("ERROR: Failed to dial connection. network=%s addr=%s error=%s\n", network, addr, err)
			return nil, err
		}

		if tcpConn, ok := conn.(*net.TCPConn); ok {
			if err := tcpConn.SetKeepAlive(true); err != nil {
				log.Println("ERROR: Failed to enable TCP KeepAlive Probes on socket:", err.Error())
				return nil, err
			}

			err := tcpConn.SetKeepAliveConfig(net.KeepAliveConfig{
				Enable:   true,
				Idle:     probeInterval,
				Interval: probeInterval,
				Count:    maxProbes,
			})
			if err != nil {
				log.Println("ERROR: Failed to configure TCP KeepAlive Probes:", err.Error())
				return nil, err
			}

			if err = tcpConn.SetNoDelay(true); err != nil {
				log.Println("ERROR: Failed to disable Nagle's algorithm for callback connection:", err.Error())
				return nil, err
			}

			log.Println("INFO: TCP keep-alive probes configured successfully for " + addr)
		}
		return conn, nil
	}

	return transport
}

func ServiceConfig(config config.HTTPClientConfig) HTTPClientConfig {
	return HTTPClientConfig{
		DialTimeout:         config.DialTimeout,
		RequestTimeout:      config.RequestTimeout,
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		ProbeInterval:       config.ProbeInterval,
		MaxProbes:           config.MaxProbes,
	}
}
