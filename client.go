package bitbuy

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const baseURL = "https://partner.bcm.exchange"

type Error struct {
	status int
	body   string
}

func (e *Error) Error() string {
	return fmt.Sprintf("bitbuy error: %d: %v", e.status, e.body)
}

type Signature struct {
	Path          string `json:"path"`
	ContentLength int    `json:"content-length"`
	Query         string `json:"query"`
}

const signatureJSON = `{"path":%s,"content-length":%s,"query":%s}`

func (s Signature) GetOrderedJSON() []byte {
	pathJson := fmt.Sprintf("\"%s\"", s.Path)
	contentLengthJson := fmt.Sprintf("%d", s.ContentLength)
	queryJson := fmt.Sprintf("\"%s\"", s.Query)

	jsonOut := fmt.Sprintf(signatureJSON, string(pathJson), contentLengthJson, queryJson)

	return []byte(jsonOut)
}

type Client interface {
	GetWallets(ctx context.Context) ([]*Wallet, error)
	Close() error
}

type client struct {
	publicKey  string
	privateKey string
}

func NewClient(publicKey string, privateKey string) Client {
	return &client{
		publicKey:  publicKey,
		privateKey: privateKey,
	}
}

func (c *client) Close() error {
	return nil
}

func (c *client) signRequest(req *http.Request) error {
	stamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	query := req.URL.Query()
	query.Add("apikey", c.publicKey)
	query.Add("stamp", stamp)
	req.URL.RawQuery = query.Encode()

	contentLength := -1
	if req.Method != http.MethodGet {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("error reading out request body: %w", err)
		}

		contentLength = len(body)
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	signaturePayload := Signature{
		Path:          req.URL.Path,
		ContentLength: contentLength,
		Query:         req.URL.RawQuery,
	}

	signatureBytes := signaturePayload.GetOrderedJSON()

	h := hmac.New(sha256.New, []byte(c.privateKey))
	h.Write(signatureBytes)
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	req.Header.Add("signature", signature)

	return nil
}

func (c *client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error during request: %w", err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		apiErr := &Error{
			status: resp.StatusCode,
			body:   string(b),
		}

		return nil, apiErr
	}

	return b, nil
}

func (c *client) getRequest(ctx context.Context, uri string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%v%v", baseURL, uri), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating get request: %w", err)
	}

	err = c.signRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error signing request: %w", err)
	}
	return c.doRequest(req)
}
