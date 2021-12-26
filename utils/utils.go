package utils

import (
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
		req.Header.Set(k, v)
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
