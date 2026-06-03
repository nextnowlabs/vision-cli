package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

func volcSignature(ak, sk, region, service, method, query string, body []byte, reqHeaders map[string]string) map[string]string {
	t := time.Now().UTC()
	dateStr := t.Format("20060102T150405Z")
	shortDate := t.Format("20060102")

	bodyHash := sha256Hex(body)

	headers := make(map[string]string)
	for k, v := range reqHeaders {
		headers[strings.ToLower(k)] = v
	}
	headers["host"] = "open.volcengineapi.com"
	headers["x-date"] = dateStr

	credentialScope := shortDate + "/" + region + "/" + service + "/request"

	signedHeaders := ""
	var signedHeaderKeys []string
	for k := range headers {
		signedHeaderKeys = append(signedHeaderKeys, k)
	}
	sort.Strings(signedHeaderKeys)
	signedHeaders = strings.Join(signedHeaderKeys, ";")

	canonicalHeaders := ""
	for _, k := range signedHeaderKeys {
		canonicalHeaders += k + ":" + strings.TrimSpace(headers[k]) + "\n"
	}

	canonicalQuery := ""
	if query != "" {
		canonicalQuery = query
	}

	canonicalRequest := strings.Join([]string{
		method,
		"/",
		canonicalQuery,
		canonicalHeaders,
		signedHeaders,
		bodyHash,
	}, "\n")

	signature := hmacSHA256Sign(sk, shortDate, region, service, canonicalRequest)

	result := map[string]string{
		"X-Date": dateStr,
	}
	result["Authorization"] = fmt.Sprintf("HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		ak, credentialScope, signedHeaders, signature)
	return result
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func hmacSHA256Sign(sk, shortDate, region, service, canonicalRequest string) string {
	kDate := hmacSHA256([]byte(sk), shortDate)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "request")

	hashedCanonical := sha256Hex([]byte(canonicalRequest))
	stringToSign := "HMAC-SHA256\n" + shortDate + "\n" + shortDate + "/" + region + "/" + service + "/request\n" + hashedCanonical

	sig := hmacSHA256(kSigning, stringToSign)
	return hex.EncodeToString(sig)
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func httpPostJSONVolc(client *http.Client, url, action, version, region, service string, query map[string]string, body map[string]any, ak, sk string) (map[string]any, error) {
	q := ""
	if len(query) > 0 {
		parts := make([]string, 0, len(query))
		for k, v := range query {
			parts = append(parts, k+"="+v)
		}
		q = strings.Join(parts, "&")
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	reqHeaders := volcSignature(ak, sk, region, service, "POST", q, jsonBody, map[string]string{
		"content-type": "application/json; charset=utf-8",
	})

	fullURL := url
	if q != "" {
		fullURL += "?" + q
	}

	req, err := http.NewRequest("POST", fullURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range reqHeaders {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
