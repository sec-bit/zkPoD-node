package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func sendRequest(body url.Values, method string, url string, Log ILogger) (string, error) {

	if method == REQUEST_METHOD_POST {
		resp, err := http.PostForm(url, body)
		if err != nil {
			Log.Errorf("connection error! url=%v, err=%v", url, err)
			return "", errors.New("connection error")
		}
		defer resp.Body.Close()

		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Log.Warnf("read response error! url=%s, err=%v", url, err)
			return string(responseBody), errors.New("read response error")
		}
		return string(responseBody), nil
	}

	client := &http.Client{Timeout: REQUEST_TIMEOUT * time.Second}

	req, err := http.NewRequest(method, url, bytes.NewReader([]byte(body.Encode())))
	if err != nil {
		Log.Errorf("create http request error. err=%v", err.Error())
		return "", errors.New("create http request error")
	}
	Log.Debugf("send request =%v, body=%v", req.Form, body)

	req.Header.Set("Content-Type", "x-www-form-urlencoded")
	req.Header.Set("Cache-Control", "no-cache")
	resp, err := client.Do(req)
	if err != nil {
		Log.Errorf("connection error! url=%v, err=%v", url, err)
		return "", errors.New("connection error")
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", url, err)
		return string(responseBody), errors.New("read response error")
	}
	return string(responseBody), nil
}
