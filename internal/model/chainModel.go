package model

type ChainItem struct {
	Item  string      `json:"item"`
	Chain []ChainItem `json:"chain,omitempty"`
}

type ChainStatement struct {
	Root      string      `json:"root"`
	Chain     []ChainItem `json:"chain,omitempty"`
	ChainName string      `json:"chainName"`
}
