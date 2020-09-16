package lotus

import (
	"encoding/json"
	"net/http"

	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
)

type Config struct {
	Url string
}

type Node struct {
	config Config
}

func New(config Config) *Node {
	return &Node{config: config}
}

func (n Node) sendRequest(method string, params ...interface{}) (string, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	}

	req := gorequest.New()
	resp, body, errs := req.Post(n.config.Url).Send(payload).End()
	if len(errs) > 0 {
		return "", errors.Wrap(errs[0], "could not execute rpc request")
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("expected status 200, got %s", resp.Status)
	}

	return body, nil
}

func (n Node) GetVersion() (string, error) {
	data, err := n.sendRequest("Filecoin.Version")
	if err != nil {
		return "", err
	}

	var versionResp Version
	err = json.Unmarshal([]byte(data), &versionResp)
	if err != nil {
		return "", errors.Wrap(err, "could not decode response from node")
	}

	return versionResp.Result.Version, nil
}
