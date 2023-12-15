package coinbase

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Client struct {
	BaseURL string
	Secret  string
	Key     string
}

func NewClient(secret, key, passphrase string) *Client {
	client := Client{
		BaseURL: "https://api.pro.coinbase.com",
		Secret:  secret,
		Key:     key,
	}

	return &client
}

func (c *Client) Request(method string, url string,
	params, result interface{}) (res *http.Response, err error) {
	var data []byte
	body := bytes.NewReader(make([]byte, 0))

	if params != nil {
		data, err = json.Marshal(params)
		if err != nil {
			return res, err
		}

		body = bytes.NewReader(data)
	}

	fullURL := fmt.Sprintf("%s%s", c.BaseURL, url)
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return res, err
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	//timestamp = "1702537003"

	// XXX: Sandbox time is off right now
	if os.Getenv("TEST_COINBASE_OFFSET") != "" {
		inc, err := strconv.Atoi(os.Getenv("TEST_COINBASE_OFFSET"))
		if err != nil {
			return res, err
		}

		timestamp = strconv.FormatInt(time.Now().Unix()+int64(inc), 10)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Baylatent Bot 2.0")
	req.Header.Add("CB-ACCESS-KEY", c.Key)
	req.Header.Add("CB-VERSION", "2015-07-22")
	req.Header.Add("CB-ACCESS-TIMESTAMP", timestamp)

	message := fmt.Sprintf("%s%s/v2%s%s", timestamp, method, url,
		string(data))

	sig, err := c.generateSig(message, c.Secret)
	if err != nil {
		return res, err
	}
	req.Header.Add("CB-ACCESS-SIGN", sig)

	client := http.Client{}
	res, err = client.Do(req)
	if err != nil {
		return res, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 && res.StatusCode != 201 {
		defer res.Body.Close()
		coinbaseError := Error{}
		decoder := json.NewDecoder(res.Body)
		if err := decoder.Decode(&coinbaseError); err != nil {
			return res, err
		}

		return res, error(coinbaseError)
	}

	if result != nil {
		decoder := json.NewDecoder(res.Body)
		if err = decoder.Decode(result); err != nil {
			return res, err
		}
	}

	return res, nil
}

func (c *Client) generateSig(message, secret string) (string, error) {
	key := []byte(secret)
	messageBytes := []byte(message)

	hash := hmac.New(sha256.New, key)
	hash.Write(messageBytes)

	digest := hash.Sum(nil)

	return hex.EncodeToString(digest), nil
}
