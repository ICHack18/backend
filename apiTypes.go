package main

type CVCategory struct {
	Name string     `json:"name"`
	Score float64   `json:"score"`
}

type CVCaption struct {
	Text string         `json:"text"`
	Confidence float64  `json:"confidence"`
}

type CVDescription struct {
	Tags []string      `json:"tags"`
	Captions []CVCaption `json:"captions"`
}
type CVResponse struct {
	Categories  []CVCategory  `json:"categories"`
	Description CVDescription `json:"description"`
}

type CVRequest struct {
	Url string  `json:"url"`
}
