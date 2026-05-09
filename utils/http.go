package utils

import (
	"crypto/tls"
	"fmt"
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

func GetBase(url string, header http.Header) ([]byte, *http.Response, error) {
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

// Get 发送 GET 请求，返回 JSON 响应
func Get[T string | map[string]any | any](url string, data T, header http.Header) (map[string]any, *http.Response, error) {
	switch v := any(data).(type) {
	case string:
		ret, response, err := GetBase(url+"?"+v, header)
		if err != nil {
			return nil, nil, err
		}
		toMap, err := ToMap(ret)
		if err != nil {
			return nil, nil, err
		}
		return toMap, response, nil
	case map[string]any:
		query := UrlEncodeMap(v)
		return Get(url, query, header)
	case any:
		toMap, err := ToMap(data)
		if err != nil {
			return nil, nil, err
		}
		return Get(url, toMap, header)
	default:
		return nil, nil, fmt.Errorf("unknown type %T", v)
	}
}
