package main

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

type ImageResponse struct {
	Url             string `json:"url"`
	Hide            bool   `json:"hide"`
	SubstituteImage string `json:"substituteImage"`
}

type Response struct {
	Images []ImageResponse `json:"images"`
}

type Request struct {
	UseCache bool  `json:"useCache"`
	Tags  []string `json:"tags"`
	Urls  []string `json:"urls"`
}

type CVCategory struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

type CVCaption struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

type CVDescription struct {
	Tags     []string    `json:"tags"`
	Captions []CVCaption `json:"captions"`
}
type CVResponse struct {
	Categories  []CVCategory  `json:"categories"`
	Description CVDescription `json:"description"`
}

type CVRequest struct {
	Url string `json:"url"`
}

type FVRequest struct {
	FaceIds []string
	PersonGroupId string
}

type Candidates struct {
	PersonId string
	Confidence float64
}

type FacesId struct {
	FaceId string
	Candidates []Candidates
}

type FVResponse struct {
	Results []FacesId
}