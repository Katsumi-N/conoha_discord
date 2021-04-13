package api

import (
	"bytes"
	"conoha/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const identityURL = "https://identity.tyo2.conoha.io/v2.0/"
const computeURL = "https://compute.tyo2.conoha.io/v2/"
const accountURL = "https://account.tyo2.conoha.io/v1/"

func doRequest(method, base string, urlPath string, tokenId string, data string, query map[string]string) (body []byte, err error) {
	client := &http.Client{}
	baseURL, err := url.Parse(base)
	if err != nil {
		return
	}
	apiURL, err := url.Parse(urlPath)
	if err != nil {
		return
	}
	// 相対パス→絶対パス
	endpoint := baseURL.ResolveReference(apiURL).String()
	log.Printf("action=doRequest endpoint=%s", endpoint)
	//リクエストの作成
	req, err := http.NewRequest(method, endpoint, bytes.NewBufferString(data))
	if err != nil {
		return
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if tokenId != "" {
		req.Header.Add("X-Auth-Token", tokenId)
	}
	// 渡されたクエリをAdd
	q := req.URL.Query()
	for key, value := range query {
		q.Add(key, value)
	}
	// クエリはエンコードが必要
	req.URL.RawQuery = q.Encode()

	// 実行
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// 帰ってきた値のbodyを読み込む
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type JsonAccess struct {
	Access JsonToken `json:"access"`
}
type JsonToken struct {
	Token TokenInfo `json:"token"`
}

type TokenInfo struct {
	Id       string `json:"id"`
	IssuedAt string `json:"issued_at"`
	Expires  string `json:"expires"`
}

// トークンの取得
func GetToken() string {
	token := TokenInfo{Id: "", IssuedAt: "", Expires: ""}
	url := "tokens"
	body := fmt.Sprintf("{\"auth\":{\"passwordCredentials\":{\"username\":\"%s\",\"password\":\"%s\"},\"tenantId\":\"%s\"}}",
		config.Config.Username, config.Config.Password, config.Config.TenantId)
	resp, err := doRequest("POST", identityURL, url, "", body, map[string]string{})
	if err != nil {
		log.Fatal(err)
	}

	var access JsonAccess
	err = json.Unmarshal(resp, &access)
	token.Id = access.Access.Token.Id
	token.IssuedAt = access.Access.Token.IssuedAt
	token.Expires = access.Access.Token.Expires

	return token.Id

}

// サーバーの各種操作(起動，シャットダウン，再起動)
func ServerCommand(tokenId string, command string) error {
	url := config.Config.TenantId + "/servers/" + config.Config.ServerId + "/action"
	var body string
	if command == "start" {
		body = fmt.Sprintf("{\"os-start\":\"null\"}")
	} else if command == "reboot" {
		body = fmt.Sprintf("{\"reboot\":{\"type\":\"SOFT\"}}")
	} else if command == "stop" {
		body = fmt.Sprintf("{\"os-stop\":\"null\"}")
	} else {
		body = ""
	}

	_, err := doRequest("POST", computeURL, url, tokenId, body, map[string]string{})
	if err != nil {
		log.Fatal(err)
	}

	return err

}

// 請求データの取得→0円と表示される
// type BillingInfo struct {
// 	BillingInvoices []JsonBillingInvoices `json:"billing_invoices"`
// }
// type JsonBillingInvoices struct {
// 	InvoiceID         int16  `json:"invoice_id"`
// 	PaymentMethodType string `json:"payment_method_type"`
// 	InvoiceDate       int    `json:"invoice_date"`
// 	BillPlasTax       int    `json:"bill_plas_tax"`
// 	DueDate           string `json:"due_date"`
// }

// func GetBilling(tokenId string, limit int) error {
// 	url := config.Config.TenantId + "/billing-invoices"
// 	query := make(map[string]string, 1)
// 	query["limit"] = strconv.Itoa(limit)
// 	resp, err := doRequest("GET", accountURL, url, tokenId, "", query)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	var bill BillingInfo
// 	err = json.Unmarshal(resp, &bill)
// 	fmt.Println("bill", bill)
// 	return err
// }

type PaymentInfo struct {
	PaymentSummary struct {
		TotalDepositAmount int `json:"total_deposit_amount"`
	} `json:"payment_summary"`
}

// 残金を出力
func GetPayment(tokenId string) (int, error) {
	url := config.Config.TenantId + "/payment-summary"

	resp, err := doRequest("GET", accountURL, url, tokenId, "", map[string]string{})
	if err != nil {
		log.Fatal(err)
	}
	var pay PaymentInfo
	err = json.Unmarshal(resp, &pay)
	return pay.PaymentSummary.TotalDepositAmount, err
}

type ServerInfo struct {
	Server struct {
		Status string `json:"status"`
	} `json:"server"`
}

// サーバーの状態を取得
func GetServerStatus(tokenId string) (status string) {
	url := config.Config.TenantId + "/servers/" + config.Config.ServerId

	resp, err := doRequest("GET", computeURL, url, tokenId, "", map[string]string{})
	if err != nil {
		log.Fatal(err)
	}
	var server ServerInfo
	err = json.Unmarshal(resp, &server)
	return server.Server.Status
}