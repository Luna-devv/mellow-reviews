package main

type Author struct {
	Id        string `json:"id"`
	Username  string `json:"username"`
	AvatarUrl string `json:"avatar_url"`
}

type Review struct {
	Author  Author `json:"author"`
	Content string `json:"content"`
	Helpful int    `json:"helpful"`
	Rating  string `json:"rating"`
}
