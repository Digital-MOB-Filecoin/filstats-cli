package lotus

type Version struct {
	Result struct {
		Version string `json:"Version"`
	} `json:"result"`
}

type PeersResp struct {
	Result []struct {
		Addrs []string
		ID    string
	} `json:"result"`
}
