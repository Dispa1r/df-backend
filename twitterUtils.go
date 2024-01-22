package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/g8rswimmer/go-twitter/v2"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

type authorize struct {
	Token string
}

func (a authorize) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

// CheckVerifyTweet
// message: public key from front-end
// pubKey: generate from message and sigBytes(the last 32 bytes)
func CheckVerifyTweet(pubKey []byte, message string) string {
	token := viper.GetString("twitter.bearKey")
	query := viper.GetString("twitter.searchTag")

	client := &twitter.Client{
		Authorizer: authorize{
			Token: token,
		},
		Client: http.DefaultClient,
		Host:   "https://api.twitter.com",
	}
	opts := twitter.TweetRecentSearchOpts{
		Expansions:  []twitter.Expansion{twitter.ExpansionEntitiesMentionsUserName, twitter.ExpansionAuthorID},
		TweetFields: []twitter.TweetField{twitter.TweetFieldCreatedAt, twitter.TweetFieldConversationID, twitter.TweetFieldAttachments},
	}

	log.Println("Callout to tweet recent search callout")

	tweetResponse, err := client.TweetRecentSearch(context.Background(), query, opts)
	if err != nil {
		log.Panicf("tweet lookup error: %v", err)
		return ""
	}

	dictionaries := tweetResponse.Raw.TweetDictionaries()

	// first find the cache
	for k, v := range userTwitterCache {
		tmpSign, err := hexutil.Decode(k)
		if err != nil {
			return ""
		}
		verified := VerifySig(message, pubKey, tmpSign)
		log.Println("cache works!", v)
		if verified {
			return v
		}
	}

	// get the twitter and cache it
	for _, v := range dictionaries {
		log.Println("author ID: ", v.Author.ID, "author Name: ", v.Author.UserName)
		log.Println("twitter text: ", v.Tweet.Text)
		tmpSign, err := GetSignFromTwitter(v.Tweet.Text)
		if err != nil {
			log.Println("fail to parse the sign bytes")
			continue
		}
		tmpSignHex := hexutil.Encode(tmpSign)
		userTwitterCache[tmpSignHex] = v.Author.UserName
		verified := VerifySig(message, pubKey, tmpSign)
		if verified {
			return v.Author.UserName
		}
	}

	return ""
}

func getUserInfo() {
	token := viper.GetString("twitter.bearKey")
	query := viper.GetString("twitter.searchTag")

	client := &twitter.Client{
		Authorizer: authorize{
			Token: token,
		},
		Client: http.DefaultClient,
		Host:   "https://api.twitter.com",
	}
	opts := twitter.TweetRecentSearchOpts{
		Expansions:  []twitter.Expansion{twitter.ExpansionEntitiesMentionsUserName, twitter.ExpansionAuthorID},
		TweetFields: []twitter.TweetField{twitter.TweetFieldCreatedAt, twitter.TweetFieldConversationID, twitter.TweetFieldAttachments},
	}

	log.Println("Callout to tweet recent search callout")

	tweetResponse, err := client.TweetRecentSearch(context.Background(), query, opts)
	if err != nil {
		log.Panicf("tweet lookup error: %v", err)
	}

	dictionaries := tweetResponse.Raw.TweetDictionaries()
	for _, v := range dictionaries {
		log.Println(v.Author.CreatedAt)

	}

}
