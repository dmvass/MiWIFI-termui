package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Stat struct {
	Devices     []DeviceStat `json:"dev"`
	Code        int          `json:"code"`
	Mem         MemStat      `json:"mem"`
	Temperature float64      `json:"temperature"`
	Count       CountStat    `json:"count"`
	Hardware    HardwareStat `json:"hardware"`
	UpTime      string       `json:"upTime"`
	CPU         CPUStat      `json:"cpu"`
	WAN         WANStat      `json:"wan"`
}

type DeviceStat struct {
	Mac              string `json:"mac"`
	MaxDownloadSpeed uint64 `json:"maxdownloadspeed,string"`
	MaxUploadSpeed   uint64 `json:"maxuploadspeed,string"`
	Upload           uint64 `json:"upload,string"`
	Download         uint64 `json:"download,string"`
	UpSpeed          uint64 `json:"upspeed,string"`
	DownSpeed        uint64 `json:"downspeed,string"`
	Online           string `json:"online"`
	Name             string `json:"devname"`
}

type MemStat struct {
	Usage float64 `json:"usage"`
	Total string  `json:"total"`
	Hz    string  `json:"hz"`
	Type  string  `json:"type"`
}

type CountStat struct {
	All    int `json:"all"`
	Online int `json:"online"`
}

type HardwareStat struct {
	Mac      string `json:"mac"`
	Platform string `json:"platform"`
	Version  string `json:"version"`
	Channel  string `json:"channel"`
	SN       string `json:"sn"`
}

type CPUStat struct {
	Core int     `json:"core"`
	Hz   string  `json:"hz"`
	Load float64 `json:"load"`
}

type WANStat struct {
	MaxDownloadSpeed uint64 `json:"maxdownloadspeed,string"`
	MaxUploadSpeed   uint64 `json:"maxuploadspeed,string"`
	Upload           uint64 `json:"upload,string"`
	Download         uint64 `json:"download,string"`
	UpSpeed          uint64 `json:"upspeed,string"`
	DownSpeed        uint64 `json:"downspeed,string"`
	Name             string `json:"devname"`
	History          string `json:"history"`
}

type Band struct {
	Manual     int     `json:"manual"`
	Code       int     `json:"code"`
	Bandwidth  float64 `json:"bandwidth"`
	Bandwidth2 float64 `json:"bandwidth2"`
	Download   float64 `json:"download"`
	Upload     float64 `json:"upload"`
}

func New(macAddress, host string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		httpClient: httpClient,
		macAddress: macAddress,
		host:       host,
	}
}

type Client struct {
	httpClient *http.Client
	macAddress string
	host       string
	token      string
}

func (c *Client) Login(username, password string) error {
	url, err := c.buildURL("/api/xqsystem/login", false)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("can't build request: %w", err)
	}

	nonce := generateNonce(c.macAddress)

	q := req.URL.Query()
	q.Add("username", username)
	q.Add("password", hashPassword(password, nonce))
	q.Add("logtype", "2")
	q.Add("nonce", nonce)
	req.URL.RawQuery = q.Encode()

	payload := struct {
		Token string `json:"token"`
	}{}

	if err := c.do(req, &payload); err != nil {
		return err
	}

	c.token = payload.Token

	return nil
}

func (c *Client) Logout() error {
	url, err := c.buildURL("/web/logout", true)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("can't build request: %w", err)
	}

	if err := c.do(req, nil); err != nil {
		return err
	}

	return nil
}

func (c *Client) Status() (Stat, error) {
	var stat Stat

	url, err := c.buildURL("/api/misystem/status", true)
	if err != nil {
		return stat, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return stat, fmt.Errorf("can't build request: %w", err)
	}

	if err := c.do(req, &stat); err != nil {
		return stat, err
	}

	return stat, nil
}

func (c *Client) BandwidthTest(history bool) (Band, error) {
	var band Band

	url, err := c.buildURL("/api/misystem/bandwidth_test", true)
	if err != nil {
		return band, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return band, fmt.Errorf("can't build request: %w", err)
	}

	if history {
		q := req.URL.Query()
		q.Add("history", "1")
		req.URL.RawQuery = q.Encode()
	}

	if err := c.do(req, &band); err != nil {
		return band, err
	}

	return band, nil
}

func (c Client) do(req *http.Request, payload interface{}) error {
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if payload == nil {
		return nil // do nothing with response
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("response body error: %w", err)
	}

	err = json.Unmarshal(body, payload)
	if err != nil {
		return fmt.Errorf("unmarshaling error: %w", err)
	}

	return nil
}

func (c *Client) buildURL(resource string, requireAuth bool) (string, error) {
	if requireAuth && c.token == "" {
		return "", errors.New("client is not authorized")
	}
	url := c.host + "/cgi-bin/luci"
	if requireAuth {
		url += "/;stok=" + c.token
	}
	return url + resource, nil
}
