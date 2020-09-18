package lotus

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
)

type Config struct {
	Url   string
	Token string
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
	if n.config.Token != "" {
		req.AppendHeader("Authorization", "Bearer "+n.config.Token)
	}

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

func (n Node) GetPeers() (int, error) {
	data, err := n.sendRequest("Filecoin.NetPeers")
	if err != nil {
		return 0, errors.Wrap(err, "could not call NetPeers")
	}

	var peersResp PeersResp
	err = json.Unmarshal([]byte(data), &peersResp)
	if err != nil {
		return 0, errors.Wrap(err, "could not decode response from NetPeers")
	}

	fmt.Println(data)

	for _, p := range peersResp.Result {
		fmt.Println(n.sendRequest("Filecoin.NetFindPeer", p.ID))
	}

	if len(peersResp.Result) > 0 {
		// fmt.Println(n.sendRequest("Filecoin.NetFindPeer", peersResp.Result[0].ID))
	}

	return len(peersResp.Result), nil
}

func (n Node) ListMiners() {
	// x.sendRequest("Filecoin.StateListMiners", nil)
	spew.Dump(n.sendRequest("Filecoin.StateListMiners", nil))
}

func (n Node) ChainHead() {
	// {
	//   "jsonrpc":"2.0",
	//   "result":{
	//      "Cids":[
	//         {
	//            "/":"bafy2bzaceawdyrbelphtqgxuroeewf566rd2rofbjavq7d5tbkergme7ide3y"
	//         }
	//      ],
	//      "Blocks":[
	//         {
	//            "Miner":"t01000",
	//            "Ticket":{
	//               "VRFProof":"suvXAWz6sVSkI2bwfEf7Jcj8u2c5QjsAayt+dpsbH5Z2O+SDBQoSl56gCfdmJSNHFGxeMxK/Q74MKK3CGgPMr3mv9J0NJX0ZzCKOtbJ/WjCAi8/q/b+jH3b4FCgH6tB9"
	//            },
	//            "ElectionProof":{
	//               "WinCount":4,
	//               "VRFProof":"iQFJzOQa14Ve6fGLtpq8aexvU/RLxCphABXkOzVU628XB2jAA2Yyku2tqSlwoEHKBTUvRmr/a3zUssWaD2JOgKpv5Y3S5EscsirB9ct2dg9snWtkIee30Zyp8TiwK24W"
	//            },
	//            "BeaconEntries":null,
	//            "WinPoStProof":[
	//               {
	//                  "PoStProof":0,
	//                  "ProofBytes":"dmFsaWQgcHJvb2Y="
	//               }
	//            ],
	//            "Parents":[
	//               {
	//                  "/":"bafy2bzacedt34nek523ywrrg2n4att6unl7w5nghgs7zqrvngbgzi4mptiklq"
	//               }
	//            ],
	//            "ParentWeight":"5796480",
	//            "Height":1007,
	//            "ParentStateRoot":{
	//               "/":"bafy2bzacedcjkjxlapowkfm62bhuenguiwaja2bpavgk4hs65kgi6wqqvvu36"
	//            },
	//            "ParentMessageReceipts":{
	//               "/":"bafy2bzacedswlcz5ddgqnyo3sak3jmhmkxashisnlpq6ujgyhe4mlobzpnhs6"
	//            },
	//            "Messages":{
	//               "/":"bafy2bzacecmda75ovposbdateg7eyhwij65zklgyijgcjwynlklmqazpwlhba"
	//            },
	//            "BLSAggregate":{
	//               "Type":2,
	//               "Data":"wAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	//            },
	//            "Timestamp":1600341240,
	//            "BlockSig":{
	//               "Type":2,
	//               "Data":"sdOio/PkKPiOK3BDym7Fgr76sl+xp35gwHHfXpuyFBPj2pVeRbwiyI5VmAbnXh4PCmNWhhQ5l7aEbcutC7IwCGy0eAi7pUIdCeuXhOyLQ9wsyr1pUTttd/CvXSVKgZyb"
	//            },
	//            "ForkSignaling":0,
	//            "ParentBaseFee":"100"
	//         }
	//      ],
	//      "Height":1007
	//   },
	//   "id":1
	// }

	spew.Dump(n.sendRequest("Filecoin.ChainHead"))
}
