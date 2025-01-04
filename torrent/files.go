// DO NOT EDIT
package torrent

type Files []struct {
	Availability float64 `json:"availability"`
	Index        int64   `json:"index"`
	IsSeed       bool    `json:"is_seed"`
	Name         string  `json:"name"`
	PieceRange   []int64 `json:"piece_range"`
	Priority     int64   `json:"priority"`
	Progress     float64 `json:"progress"`
	Size         int64   `json:"size"`
}
