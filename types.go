package main

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

type Response struct {
	Hide            bool   `json:"hide"`
	SubstituteImage string `json:"substituteImage"`
}

type Request struct {
	Cache    bool     `json:"cache"`
	Tags     []string `json:"tags"`
	ImageURL string   `json:"imageURL"`
	Image    []byte   `json:"image"`
}

type Category struct {
	Name string     `json:"name"`
	Score float64   `json:"score"`
}
type MSResponse struct {
	Categories  []Category  `json:"categories"`
}

type MSRequest struct {
	Url string  `json:"url"`
}
