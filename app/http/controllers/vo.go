package controllers

type LinkItem struct {
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Url     string `json:"url"`
	LogoUrl string `json:"logoUrl"`
}
