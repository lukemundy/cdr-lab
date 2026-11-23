package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// The media types of a multi-arch index and of a single image, so that the registry finds each form of a tag.
const manifestTypes = "application/vnd.oci.image.index.v1+json, application/vnd.oci.image.manifest.v1+json, " +
	"application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.docker.distribution.manifest.v2+json"

// registry reads a public image repository with the OCI distribution API and an anonymous pull token.
type registry struct {
	client *http.Client
	host   string
	name   string
	token  string
}

func newRegistry(client *http.Client, image string) (*registry, error) {
	host, name, ok := strings.Cut(image, "/")
	if !ok || name == "" {
		return nil, fmt.Errorf("the image %q has no registry host and repository name", image)
	}

	return &registry{client: client, host: host, name: name}, nil
}

func (r *registry) has(ctx context.Context, tag string) (bool, error) {
	status, challenge, err := r.head(ctx, tag)
	if err != nil {
		return false, err
	}

	if status == http.StatusUnauthorized && r.token == "" {
		if err := r.getToken(ctx, challenge); err != nil {
			return false, err
		}

		if status, _, err = r.head(ctx, tag); err != nil {
			return false, err
		}
	}

	switch status {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("get the manifest %s/%s:%s: HTTP status %d", r.host, r.name, tag, status)
	}
}

func (r *registry) head(ctx context.Context, tag string) (status int, challenge string, err error) {
	u := url.URL{Scheme: "https", Host: r.host, Path: "/v2/" + r.name + "/manifests/" + tag}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, u.String(), http.NoBody)
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Accept", manifestTypes)

	if r.token != "" {
		req.Header.Set("Authorization", "Bearer "+r.token)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("get the manifest %s/%s:%s: %w", r.host, r.name, tag, err)
	}

	_ = resp.Body.Close()

	return resp.StatusCode, resp.Header.Get("WWW-Authenticate"), nil
}

func (r *registry) getToken(ctx context.Context, challenge string) error {
	params, ok := strings.CutPrefix(challenge, "Bearer ")
	if !ok {
		return fmt.Errorf("the registry %s gives no Bearer challenge", r.host)
	}

	values := map[string]string{}

	for part := range strings.SplitSeq(params, ",") {
		key, value, _ := strings.Cut(strings.TrimSpace(part), "=")
		values[key] = strings.Trim(value, `"`)
	}

	realm, err := url.Parse(values["realm"])
	if err != nil || realm.Scheme != "https" {
		return fmt.Errorf("the registry %s gives a token realm that is not an HTTPS URL", r.host)
	}

	query := realm.Query()
	query.Set("service", values["service"])
	query.Set("scope", values["scope"])
	realm.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, realm.String(), http.NoBody)
	if err != nil {
		return err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("get a pull token from %s: %w", r.host, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get a pull token from %s: %s", r.host, resp.Status)
	}

	var body struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.Token == "" {
		return errors.Join(fmt.Errorf("the registry %s gives no pull token", r.host), err)
	}

	r.token = body.Token

	return nil
}
