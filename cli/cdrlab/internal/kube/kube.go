// Package kube reads and writes the Secrets of one namespace through the Kubernetes API, as the ServiceAccount of the
// pod. It uses no Kubernetes client library, because the commands make only these calls.
package kube

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const TLSType = "kubernetes.io/tls"

// ServiceAccountDir is where Kubernetes mounts the token, the CA certificate and the namespace of the ServiceAccount.
const ServiceAccountDir = "/var/run/secrets/kubernetes.io/serviceaccount"

const (
	requestTimeout = 30 * time.Second
	maxBody        = 1 << 20
)

type Secret struct {
	Name string
	Type string
	Data map[string][]byte
	// ResourceVersion is the version that Get gave. Update refuses a Secret that changed after it.
	ResourceVersion string
}

type Client struct {
	base      string
	namespace string
	tokenFile string
	http      *http.Client
}

func New(baseURL, namespace, tokenFile string, httpClient *http.Client) *Client {
	return &Client{base: strings.TrimSuffix(baseURL, "/"), namespace: namespace, tokenFile: tokenFile, http: httpClient}
}

// InCluster gives a client for the namespace of the pod, with the ServiceAccount files in dir.
func InCluster(dir string) (*Client, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" || port == "" {
		return nil, errors.New(
			"KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT are not set: run the command in a pod",
		)
	}

	//nolint:gosec // G304: dir is the directory of the ServiceAccount files
	namespace, err := os.ReadFile(filepath.Join(dir, "namespace"))
	if err != nil {
		return nil, fmt.Errorf("read the namespace of the pod: %w", err)
	}

	//nolint:gosec // G304: dir is the directory of the ServiceAccount files
	caPEM, err := os.ReadFile(filepath.Join(dir, "ca.crt"))
	if err != nil {
		return nil, fmt.Errorf("read the CA certificate of the Kubernetes API: %w", err)
	}

	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caPEM) {
		return nil, errors.New("the CA certificate of the Kubernetes API is not PEM")
	}

	httpClient := &http.Client{
		Timeout:   requestTimeout,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}},
	}

	return New(
		"https://"+net.JoinHostPort(host, port),
		strings.TrimSpace(string(namespace)),
		filepath.Join(dir, "token"),
		httpClient,
	), nil
}

func (c *Client) Namespace() string { return c.namespace }

type metadata struct {
	Name            string `json:"name"`
	Namespace       string `json:"namespace,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

type object struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   metadata          `json:"metadata"`
	Type       string            `json:"type,omitempty"`
	Data       map[string][]byte `json:"data,omitempty"`
}

// Get gives found false, and no error, when the Secret does not exist.
func (c *Client) Get(ctx context.Context, name string) (s Secret, found bool, err error) {
	var o object

	status, err := c.do(ctx, http.MethodGet, name, nil, &o)

	switch {
	case status == http.StatusNotFound:
		return Secret{}, false, nil
	case err != nil:
		return Secret{}, false, fmt.Errorf("get the Secret %q: %w", name, err)
	}

	return Secret{
		Name:            o.Metadata.Name,
		Type:            o.Type,
		Data:            o.Data,
		ResourceVersion: o.Metadata.ResourceVersion,
	}, true, nil
}

func (c *Client) Create(ctx context.Context, s Secret) error {
	if _, err := c.do(ctx, http.MethodPost, "", c.object(s), nil); err != nil {
		return fmt.Errorf("create the Secret %q: %w", s.Name, err)
	}

	return nil
}

func (c *Client) Update(ctx context.Context, s Secret) error {
	if _, err := c.do(ctx, http.MethodPut, s.Name, c.object(s), nil); err != nil {
		return fmt.Errorf("update the Secret %q: %w", s.Name, err)
	}

	return nil
}

func (c *Client) Delete(ctx context.Context, name string) error {
	if _, err := c.do(ctx, http.MethodDelete, name, nil, nil); err != nil {
		return fmt.Errorf("delete the Secret %q: %w", name, err)
	}

	return nil
}

func (c *Client) object(s Secret) object {
	return object{
		APIVersion: "v1",
		Kind:       "Secret",
		Metadata:   metadata{Name: s.Name, Namespace: c.namespace, ResourceVersion: s.ResourceVersion},
		Type:       s.Type,
		Data:       s.Data,
	}
}

// do gives the status of the response, also with an error for a status that is not 2xx.
func (c *Client) do(ctx context.Context, method, name string, body, out any) (int, error) {
	target := c.base + "/api/v1/namespaces/" + url.PathEscape(c.namespace) + "/secrets"
	if name != "" {
		target += "/" + url.PathEscape(name)
	}

	var reader io.Reader = http.NoBody

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}

		reader = bytes.NewReader(data)
	}

	token, err := os.ReadFile(c.tokenFile)
	if err != nil {
		return 0, fmt.Errorf("read the ServiceAccount token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return resp.StatusCode, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp.StatusCode, statusError(resp, data)
	}

	if out == nil {
		return resp.StatusCode, nil
	}

	return resp.StatusCode, json.Unmarshal(data, out)
}

// statusError gives the message of the Status object of the API, which names the object but has none of its data.
func statusError(resp *http.Response, data []byte) error {
	var status struct {
		Message string `json:"message"`
	}

	if json.Unmarshal(data, &status) != nil || status.Message == "" {
		return fmt.Errorf("the Kubernetes API gave %s", resp.Status)
	}

	return fmt.Errorf("the Kubernetes API gave %s: %s", resp.Status, status.Message)
}
