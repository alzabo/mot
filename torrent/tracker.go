package torrent

type Tracker struct {
	Hash          string `json:"hash,omitempty"`
	Msg           string `json:"msg"`
	NumDownloaded int64  `json:"num_downloaded"`
	NumLeeches    int64  `json:"num_leeches"`
	NumPeers      int64  `json:"num_peers"`
	NumSeeds      int64  `json:"num_seeds"`
	Status        int64  `json:"status"`
	Tier          int64  `json:"tier"`
	URL           string `json:"url"`
}
