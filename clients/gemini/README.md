[![GitHub license](https://claudiocandio.github.io/img/license_mit.svg)](https://github.com/claudiocandio/gemini-api/blob/master/LICENSE)
[![Language: Go](https://claudiocandio.github.io/img/language-Go.svg)](https://golang.org/)
[![Donate Bitcoin](https://claudiocandio.github.io/img/donate-bitcoin-orange.svg)](https://claudiocandio.github.io/img/donate-bitcoin.html)
[![Donate Ethereum](https://claudiocandio.github.io/img/donate-etherum-green.svg)](https://claudiocandio.github.io/img/donate-ethereum.html)

# API wrapper for the Gemini Exchange REST API

Gemini-api is a wrapper for the Gemini Exchange REST API <https://docs.gemini.com/rest-api/> it can connect to the production site <https://api.gemini.com> or to the Sandbox site <https://api.sandbox.gemini.com> for testing purposes.

This package is based on the <https://github.com/jsgoyette/gemini>, I have rewritten many things, added some more functionalities, debug & trace logging but I haven't had time to fully test everything, it is working fine for me but use it at your own risk.

## gemini_cli

This package is used by the gemini_cli <https://github.com/claudiocandio/gemini_cli> a cli command to facilitate the use of the Gemini Exchange via REST API.

## Usage Example

```bash
$ go get github.com/claudiocandio/gemini-api
```

Simple example:

```golang
package main

import (
 "encoding/json"
 "fmt"

 "github.com/claudiocandio/gemini-api"
)

func main() {

 api := gemini.New(
  false, // if this is false, it will use Gemini Sandox site: <https://api.sandbox.gemini.com>
         // if this is true,  it will use Gemini Production site: <https://api.gemini.com>
  "MyGeminiApiKey",    // GEMINI_API_KEY
  "MyGeminiApiSecret", // GEMINI_API_SECRET
 )

 // check more api methods in private.go & public.go
 accountDetail, err := api.AccountDetail()
 if err != nil {
  fmt.Printf("Error AccountDetail: %s\n", err)
  return
 }
 j, err := json.MarshalIndent(&accountDetail, "", " ")
 if err != nil {
  fmt.Printf("Error MarshalIndent: %s\n", err)
 }

 fmt.Printf("%s", j)
}
```

Check more api methods in private.go & public.go

## Disclaimer

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND.
