package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
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

func PostBase(url string, header http.Header, data []byte) ([]byte, *http.Response, error) {
	buffer := bytes.NewBuffer(data)
	h := NewHttp()
	req, err := http.NewRequest("POST", url, buffer)
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

type TypePost uint

const (
	PoseDefault TypePost = iota
	PoseKV
	PoseFormData
)

func Post[T []byte | string | map[string]any | any](url string, header http.Header, data T, types TypePost) ([]byte, *http.Response, error) {
	switch v := any(data).(type) {
	case []byte:
		return PostBase(url, header, v)
	case string:
		return PostBase(url, header, []byte(v))
	case map[string]any:
		switch types {
		case PoseDefault:
			json, err := ToJson(v)
			if err != nil {
				return nil, nil, err
			}
			return PostBase(url, header, json)
		case PoseKV:
			encodeMap := UrlEncodeMap(v)
			return PostBase(url, header, []byte(encodeMap))
		case PoseFormData:
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			for kk, vv := range v {
				switch vvv := vv.(type) {
				case string:
					writer.WriteField(kk, vvv)
				default:
					return nil, nil, fmt.Errorf("unknown type %T", vv)
				}
			}
			err := writer.Close()
			if err != nil {
				return nil, nil, err
			}
			return PostBase(url, header, []byte(body.String()))
		default:
			return nil, nil, fmt.Errorf("unknown type %T", v)
		}
	case any:
		toMap, err := ToMap(v)
		if err != nil {
			return nil, nil, err
		}
		return Post(url, header, toMap, types)
	default:
		return nil, nil, fmt.Errorf("unknown type %T", v)
	}
}
