package model

type MapConfig struct {
	TencentMapKey string `json:"tencent_map_key"`
}

type Location struct {
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}