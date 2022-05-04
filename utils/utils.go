package utils

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
)

func RandString(n int) string {
	key := make([]byte, n)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%x", key)
}

func Post(url, data string, header map[string]string) ([]byte, error) {
	req, _ := http.NewRequest("POST", url, strings.NewReader(data))
	for k, v := range header {
		req.Header.Add(k, v)
	}
	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func DecodeAuth(auth string) (string, string) {
	var clientID, clientSecret string
	if strings.HasPrefix(auth, "Basic ") {
		s := auth[len("basic "):]
		list := strings.Split(s, ":")
		if len(list) > 1 {
			var secretBase64 string
			clientID, secretBase64 = list[0], list[1]
			b, _ := base64.StdEncoding.DecodeString(secretBase64)
			clientSecret = string(b)
		}
	}
	return clientID, clientSecret
}