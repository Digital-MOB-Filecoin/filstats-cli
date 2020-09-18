package lotus

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	var x *Node

	x = New(Config{
		Url: "http://localhost:8000/rpc/v0",
	})

	// x.GetPeers()

	// fmt.Println(x.sendRequest("Filecoin.StateListMiners", nil))
	// fmt.Println(x.sendRequest("Filecoin.ChainHead"))
	fmt.Println(x.sendRequest("Filecoin.StateMinerPower", "t01000", nil))
}
