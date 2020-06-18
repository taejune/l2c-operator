package utils

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
)

func GetSonarServerAccessInfo() (url string, token string) {
	url = os.Getenv("SONAR_URL")
	token = os.Getenv("SONAR_TOKEN")

	if url == "" {
		url = "http://sonarqube-test-deploy-service.default:9000"
	}

	if token == "" {
		token = "29dc730637516e2e86b52e30f0693e83b6af6513"
	}

	return url, token
}

func CreateProject() {}

func postReq(uri string, body interface{}, contentType string) (resp string, err error) {
	var rawResp *http.Response

	if contentType == "" {
		contentType = "text/plain"
	}

	switch body.(type) {
	case string:
		rawResp, err = http.Post(uri, contentType, bytes.NewBufferString(body.(string)))
		break
	case url.Values:
		rawResp, err = http.PostForm(uri, body.(url.Values))
		break
	default:
		return "", errors.New("Not supported type " + reflect.TypeOf(body).String())
	}

	if err != nil || rawResp == nil {
		return "", err
	}

	respBytes, err := ioutil.ReadAll(rawResp.Body)
	if err != nil {
		return "", nil
	}
	resp = string(respBytes)

	if err = rawResp.Body.Close(); err != nil {
		return "", nil
	}

	return resp, err
}
