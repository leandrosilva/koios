package main

// Movie record in the Netflix dataset
type Movie struct {
	ShowID      int    `json:"show_id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Director    string `json:"director"`
	Cast        string `json:"cast"`
	Country     string `json:"country"`
	DateAdded   string `json:"date_added"`
	ReleaseYear int    `json:"release_year"`
	Rating      string `json:"rating"`
	Duration    string `json:"duration"`
	ListedIn    string `json:"listed_in"`
	Description string `json:"description"`
}
