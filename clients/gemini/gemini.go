package gemini

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/claudiocandio/gemini-api/logger"
)

type Api struct {
	url    string
	key    string
	secret string
}

// buildHeader handles the conversion of post parameters into headers formatted
// according to Gemini specification. Resulting headers include the API key,
// the payload and the signature.
func (api *Api) buildHeader(req *map[string]interface{}) http.Header {

	reqStr, _ := json.Marshal(req)
	payload := base64.StdEncoding.EncodeToString([]byte(reqStr))

	mac := hmac.New(sha512.New384, []byte(api.secret))
	if _, err := mac.Write([]byte(payload)); err != nil {
		panic(err)
	}

	signature := hex.EncodeToString(mac.Sum(nil))

	header := http.Header{}
	header.Set("X-GEMINI-APIKEY", api.key)
	header.Set("X-GEMINI-PAYLOAD", payload)
	header.Set("X-GEMINI-SIGNATURE", signature)

	return header
}

// request makes the HTTP request to Gemini and handles any returned errors
func (api *Api) request(verb, url string, params map[string]interface{}) ([]byte, error) {

	logger.Debug("func request: http.NewRequest",
		fmt.Sprintf("verb:%s", verb),
		fmt.Sprintf("url:%s", url),
		fmt.Sprintf("params:%v", params),
	)

	req, err := http.NewRequest(verb, url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}

	if params != nil {
		if verb == "GET" {
			q := req.URL.Query()
			for key, val := range params {
				q.Add(key, val.(string))
			}
			req.URL.RawQuery = q.Encode()
		} else {
			req.Header = api.buildHeader(&params)
		}
	}

	// this will also show gemini key and secret, pay attention
	logger.Trace("func request: params",
		fmt.Sprintf("req:%v", req),
	)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	logger.Debug("func request: Http Client response",
		fmt.Sprintf("resp:%v", resp),
	)

	// read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	//unsuccessful status code
	if resp.StatusCode > 299 {
		statusCode := fmt.Sprintf("HTTP Status Code: %d", resp.StatusCode)
		return nil, fmt.Errorf("%s\n%s", statusCode, body)
	}

	logger.Debug("func request: Http Client body",
		fmt.Sprintf("body:%s", body),
	)

	return body, nil
}

func New(live bool, key, secret string) *Api {
	var url string
	if url = sandbox_URL; live {
		url = base_URL
	}

	return &Api{url: url, key: key, secret: secret}
}
