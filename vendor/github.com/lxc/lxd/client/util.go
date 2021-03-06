package lxd

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/lxc/lxd/shared"
)

func tlsHTTPClient(client *http.Client, tlsClientCert string, tlsClientKey string, tlsCA string, tlsServerCert string, insecureSkipVerify bool, proxy func(req *http.Request) (*url.URL, error)) (*http.Client, error) {
	// Get the TLS configuration
	tlsConfig, err := shared.GetTLSConfigMem(tlsClientCert, tlsClientKey, tlsCA, tlsServerCert, insecureSkipVerify)
	if err != nil {
		return nil, err
	}

	// Define the http transport
	transport := &http.Transport{
		TLSClientConfig:   tlsConfig,
		Dial:              shared.RFC3493Dialer,
		Proxy:             shared.ProxyFromEnvironment,
		DisableKeepAlives: true,
	}

	// Allow overriding the proxy
	if proxy != nil {
		transport.Proxy = proxy
	}

	// Define the http client
	if client == nil {
		client = &http.Client{}
	}
	client.Transport = transport

	// Setup redirect policy
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// Replicate the headers
		req.Header = via[len(via)-1].Header

		return nil
	}

	return client, nil
}

func unixHTTPClient(client *http.Client, path string) (*http.Client, error) {
	// Setup a Unix socket dialer
	unixDial := func(network, addr string) (net.Conn, error) {
		raddr, err := net.ResolveUnixAddr("unix", path)
		if err != nil {
			return nil, err
		}

		return net.DialUnix("unix", nil, raddr)
	}

	// Define the http transport
	transport := &http.Transport{
		Dial:              unixDial,
		DisableKeepAlives: true,
	}

	// Define the http client
	if client == nil {
		client = &http.Client{}
	}
	client.Transport = transport

	// Setup redirect policy
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// Replicate the headers
		req.Header = via[len(via)-1].Header

		return nil
	}

	return client, nil
}

type nullReadWriteCloser int

func (nullReadWriteCloser) Close() error                { return nil }
func (nullReadWriteCloser) Write(p []byte) (int, error) { return len(p), nil }
func (nullReadWriteCloser) Read(p []byte) (int, error)  { return 0, io.EOF }

func remoteOperationError(msg string, errors map[string]error) error {
	// Check if empty
	if len(errors) == 0 {
		return nil
	}

	// Check if all identical
	var err error
	for _, entry := range errors {
		if err != nil && entry.Error() != err.Error() {
			errorStrs := []string{}
			for server, errorStr := range errors {
				errorStrs = append(errorStrs, fmt.Sprintf("%s: %s", server, errorStr))
			}

			return fmt.Errorf("%s:\n - %s", msg, strings.Join(errorStrs, "\n - "))
		}

		err = entry
	}

	// Check if successful
	if err != nil {
		return fmt.Errorf("%s: %s", msg, err)
	}

	return nil
}
