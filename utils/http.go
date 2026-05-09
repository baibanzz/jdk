package utils

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
)

// NewHttp 创建 Http 客户端，跳过 SSL 验证
func NewHttp() http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return http.Client{Transport: tr}
}

func Get(url string, header http.Header) ([]byte, *http.Response, error) {
	h := NewHttp()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}
	for k, v := range header {
		req.Header[k] = v
	}
	resp, err := h.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}
	return body, resp, nil
}

//// Get 发送 GET 请求，返回 JSON 响应
//func GetMap[T string | map[string]any | struct{}](url string, data T, header http.Header) (map[string]any, *http.Response, error) {
//	var pushData string
//
//}

// GetToStruct 发送 GET 请求，将响应映射到结构体
func GetToStruct[T any](url string) (T, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: tr}

	var result T
	resp, err := client.Get(url)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body, &result)
	return result, err
}
