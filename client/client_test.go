package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Login(t *testing.T) {

	var (
		username = "admin"
		password = "admin"
		nonce    = generateNonce("00:11:22:33:44:55")
	)

	t.Run("ok", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected headers
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/api/xqsystem/login")
			// Expected query params
			assert.Equal(t, username, r.URL.Query().Get("username"))
			assert.Equal(t, hashPassword(password, nonce), r.URL.Query().Get("password"))
			assert.Equal(t, "2", r.URL.Query().Get("logtype"))
			assert.Equal(t, nonce, r.URL.Query().Get("nonce"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(`{"token": "token"}`))
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      nonce,
		}

		assert.NoError(t, c.Login(username, password))
		assert.Equal(t, "token", c.token)
	})

	t.Run("invalid response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected headers
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/api/xqsystem/login")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(`{}`))
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      nonce,
		}

		assert.Error(t, c.Login(username, password), "invalid token")
	})

	t.Run("client error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected headers
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/api/xqsystem/login")

			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      nonce,
		}

		assert.Error(t, c.Login(username, password))
	})

	t.Run("server error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected headers
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/api/xqsystem/login")

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      nonce,
		}

		assert.Error(t, c.Login(username, password))
	})
}

func TestClient_Logout(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/;stok=token/web/logout")

			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(200)
			w.Write([]byte(`<!DOCTYPE html>`))
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      "nonce",
			token:      "token",
		}

		assert.NoError(t, c.Logout())
	})

	t.Run("not authorized", func(t *testing.T) {
		c := Client{
			httpClient: http.DefaultClient,
			host:       "localhost",
			nonce:      "nonce",
		}

		assert.Error(t, c.Logout(), "client is not authorized")
	})

	t.Run("client error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/;stok=token/web/logout")

			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      "nonce",
			token:      "token",
		}

		assert.Error(t, c.Logout())
	})

	t.Run("server error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/;stok=token/web/logout")

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      "nonce",
			token:      "token",
		}

		assert.Error(t, c.Logout())
	})
}

func TestClient_Status(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/;stok=token/api/misystem/status")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(`
				{
					"dev": [
						{
							"mac": "00:11:22:33:44:55",
							"maxdownloadspeed": "408483",
							"upload": "88311434",
							"upspeed": "0",
							"downspeed": "0",
							"online": "168914",
							"devname": "client_1",
							"maxuploadspeed": "103678",
							"download": "1878126530"
						},
						{
							"mac": "55:44:33:22:11:00",
							"maxdownloadspeed": "414979",
							"upload": "70794013",
							"upspeed": "0",
							"downspeed": "0",
							"online": "76386",
							"devname": "client_2",
							"maxuploadspeed": "327776",
							"download": "1303317669"
						}
					],
					"code": 0,
					"mem": {
						"usage": 0.42,
						"total": "64MB",
						"hz": "800MHz",
						"type": "DDR2"
					},
					"temperature": 0,
					"count": {
						"all": 3,
						"online": 2
					},
					"hardware": {
						"mac": "AA:BB:CC:DD:EE:FF",
						"platform": "R4CM",
						"version": "3.0.16",
						"channel": "release",
						"sn": "25091/A9UT41212"
					},
					"upTime": "438770.04",
					"cpu": {
						"core": 1,
						"hz": "575MHz",
						"load": 0.7525
					},
					"wan": {
						"downspeed": "4137",
						"maxdownloadspeed": "533653",
						"history": "0,200829,180511,239543,259868,429",
						"devname": "eth0.2",
						"upload": "2320806776",
						"upspeed": "1937",
						"maxuploadspeed": "423542",
						"download": "19771478107"
					}
				}
			`))
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      "nonce",
			token:      "token",
		}

		stat, err := c.Status()
		assert.NoError(t, err)
		assert.Equal(t, Stat{
			Devices: []DeviceStat{
				{
					Mac:              "00:11:22:33:44:55",
					MaxDownloadSpeed: 408483,
					Upload:           88311434,
					UpSpeed:          0,
					DownSpeed:        0,
					Online:           "168914",
					Name:             "client_1",
					MaxUploadSpeed:   103678,
					Download:         1878126530,
				},
				{
					Mac:              "55:44:33:22:11:00",
					MaxDownloadSpeed: 414979,
					Upload:           70794013,
					UpSpeed:          0,
					DownSpeed:        0,
					Online:           "76386",
					Name:             "client_2",
					MaxUploadSpeed:   327776,
					Download:         1303317669,
				},
			},
			Code: 0,
			Mem: MemStat{
				Usage: 0.42,
				Total: "64MB",
				Hz:    "800MHz",
				Type:  "DDR2",
			},
			Temperature: 0,
			Count: CountStat{
				All:    3,
				Online: 2,
			},
			Hardware: HardwareStat{
				Mac:      "AA:BB:CC:DD:EE:FF",
				Platform: "R4CM",
				Version:  "3.0.16",
				Channel:  "release",
				SN:       "25091/A9UT41212",
			},
			UpTime: "438770.04",
			CPU: CPUStat{
				Core: 1,
				Hz:   "575MHz",
				Load: 0.7525,
			},
			WAN: WANStat{
				DownSpeed:        4137,
				MaxDownloadSpeed: 533653,
				History:          "0,200829,180511,239543,259868,429",
				Name:             "eth0.2",
				Upload:           2320806776,
				UpSpeed:          1937,
				MaxUploadSpeed:   423542,
				Download:         19771478107,
			},
		}, stat)
	})

	t.Run("client error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/;stok=token/api/misystem/status")

			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      "nonce",
			token:      "token",
		}

		_, err := c.Status()
		assert.Error(t, err)
	})

	t.Run("server error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/;stok=token/api/misystem/status")

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      "nonce",
			token:      "token",
		}

		_, err := c.Status()
		assert.Error(t, err)
	})
}

func TestClient_BandwidthTest(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/;stok=token/api/misystem/bandwidth_test")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(`
				{
					"manual": 0,
					"bandwidth2": 0.12,
					"code": 0,
					"upload": 15.36,
					"download": 147.2,
					"bandwidth": 1.15
				}
			`))
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      "nonce",
			token:      "token",
		}

		band, err := c.BandwidthTest(false)
		assert.NoError(t, err)
		assert.Equal(t, Band{
			Manual:     0,
			Code:       0,
			Bandwidth:  1.15,
			Bandwidth2: 0.12,
			Download:   147.2,
			Upload:     15.36,
		}, band)
	})

	t.Run("ok/history", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/;stok=token/api/misystem/bandwidth_test")
			// Expected query params
			assert.Equal(t, "1", r.URL.Query().Get("history"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(`{}`))
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      "nonce",
			token:      "token",
		}

		_, err := c.BandwidthTest(true)
		assert.NoError(t, err)
	})

	t.Run("client error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/;stok=token/api/misystem/bandwidth_test")

			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      "nonce",
			token:      "token",
		}

		_, err := c.BandwidthTest(false)
		assert.Error(t, err)
	})

	t.Run("server error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expected path
			assert.Equal(t, r.URL.Path, "/cgi-bin/luci/;stok=token/api/misystem/bandwidth_test")

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		c := Client{
			httpClient: http.DefaultClient,
			host:       ts.URL,
			nonce:      "nonce",
			token:      "token",
		}

		_, err := c.BandwidthTest(false)
		assert.Error(t, err)
	})
}
