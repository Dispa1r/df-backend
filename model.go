package main

type twitterMsg struct {
	sender  string
	twitter string
	sign    string
}

// locationId,
//
//	sender: this.account,
//	type,
//	body,
type emojiMsg struct {
	locationId string
	emoji      string
	sender     string
	sign       string
	msgType    string
}

type planetMessage struct {
	Id          string            `json:"id"`
	Sender      string            `json:"sender"`
	TimeCreated int64             `json:"timeCreated"`
	PlanetId    string            `json:"planetId"`
	Body        map[string]string `json:"body"`
	EmojiFlag   string            `json:"type"`
}
