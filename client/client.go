package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/cmars/peerless"
)

type Client struct {
	url     string
	auth    *peerless.Authorization
	counter uint64
}

func New(url string) *Client {
	return &Client{
		url: url,
	}
}

func (c *Client) Do() error {
	if c.auth == nil {
		return c.doAuth()
	}
	req, err := http.NewRequest("GET", strings.TrimRight(c.url, "/")+"/", nil)
	if err != nil {
		return errors.Wrap(err, "failed to create http.Request")
	}
	authJson, err := json.Marshal(c.auth)
	if err != nil {
		return errors.Wrap(err, "failed to marshal auth")
	}
	authEnc := base64.StdEncoding.EncodeToString(authJson)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", authEnc))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to make request")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		c.counter++
		c.auth.Next(c.counter)
		return nil
	} else if resp.StatusCode == http.StatusUnauthorized {
		return c.doAuth()
	} else {
		var respBody [1024]byte
		_, err := resp.Body.Read(respBody[:])
		if err != nil {
			log.Printf("failed to read response: %+v", err)
			return errors.Errorf("request failed: %s", resp.Status)
		}
		return errors.Errorf("request failed: %s %v", resp.Status, string(respBody[:]))
	}
	panic("unreachable")
}

func (c *Client) doAuth() error {
	req, err := http.NewRequest("POST", strings.TrimRight(c.url, "/")+"/token", nil)
	if err != nil {
		return errors.Wrap(err, "failed to create http.Request")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to make request")
	}
	defer resp.Body.Close()
	var respBody [1024]byte

	if resp.StatusCode != http.StatusOK {
		_, err := resp.Body.Read(respBody[:])
		if err != nil {
			log.Printf("failed to read response: %v", err)
			return errors.Errorf("request failed: %s", resp.Status)
		}
		return errors.Errorf("request failed: %s %v", resp.Status, string(respBody[:]))
	}

	n, err := resp.Body.Read(respBody[:])
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "failed to read token response")
	}
	authBytes, err := base64.StdEncoding.DecodeString(string(respBody[:n]))
	if err != nil {
		return errors.Wrap(err, "failed to decode token response")
	}
	var auth peerless.Authorization
	err = json.Unmarshal(authBytes, &auth)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal token")
	}
	c.auth = &auth
	return nil
}
