package apiclient

type City struct {
	ID   string
	Name string
}

type Movie struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

type Seance struct {
	ID       int    `json:"id,omitempty"`
	Price    int    `json:"price"`
	DateTime string `json:"datetime"`
	Seats    []Seat `json:"seats,omitempty"`
}

type Seat struct {
	ID     int  `json:"id"`
	Vacant bool `json:"vacant"`
}
