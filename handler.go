package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	"log"
	"strings"
	"time"
)

func twitterVerifyHandler(c *gin.Context) {
	buf := make([]byte, 1024)
	c.Request.Body.Read(buf)
	if !strings.HasPrefix(string(buf), "{\"verifyMessage") {
		c.JSON(500, gin.H{
			"success": false,
			"message": "invalid binding message",
		})
		return
	}

	// parse the json
	tmpMsg, err := parseTwitterMsg(buf, "bind")
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"message": err,
		})
		return
	}

	// check eth address
	fmtMsg := fmt.Sprintf("{\"twitter\":\"%s\"}", tmpMsg.twitter)
	realAddress, fullPubkey, err := generatePubKeyFromSign(tmpMsg.sign, fmtMsg)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"success": false,
			"message": err,
		})
		return
	}
	if hexutil.Encode(realAddress) != tmpMsg.sender {
		c.JSON(500, gin.H{
			"success": false,
			"message": "invalid binding message",
		})
		return
	}

	tmpSign, _ := hexutil.Decode(tmpMsg.sign)
	userCheck := VerifySig(fmtMsg, fullPubkey, tmpSign)
	if !userCheck {
		log.Println("message from invalid user")
		c.JSON(500, gin.H{
			"success": false,
			"message": err,
		})
		return
	}

	verifyResult := CheckVerifyTweet(fullPubkey, tmpMsg.twitter)
	if len(verifyResult) != 0 {
		userTwitterMap[tmpMsg.sender] = verifyResult
		c.JSON(200, gin.H{
			"success": true,
			"message": "success to bind the address",
		})
		return
	}
}

func disconnectHandler(c *gin.Context) {
	buf := make([]byte, 1024)
	c.Request.Body.Read(buf)
	if !strings.HasPrefix(string(buf), "{\"disconnectMessage") {
		c.JSON(500, gin.H{
			"success": false,
			"message": "invalid disconnect message",
		})
		return
	}

	// parse the json
	tmpMsg, err := parseTwitterMsg(buf, "disconnect")
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"message": err,
		})
		return
	}

	fmtMsg := fmt.Sprintf("{\"twitter\":\"%s\"}", tmpMsg.twitter)
	realAddress, fullPubKey, err := generatePubKeyFromSign(tmpMsg.sign, fmtMsg)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"success": false,
			"message": err,
		})
		return
	}

	if hexutil.Encode(realAddress) != tmpMsg.sender {
		c.JSON(500, gin.H{
			"success": false,
			"message": "invalid sender",
		})
		return
	}

	tmpSign, _ := hexutil.Decode(tmpMsg.sign)
	userCheck := VerifySig(fmtMsg, fullPubKey, tmpSign)
	if !userCheck {
		c.JSON(500, gin.H{
			"success": false,
			"message": err,
		})
		return
	}

	delete(userTwitterMap, tmpMsg.sender)
	c.JSON(200, gin.H{
		"success": true,
		"message": "success to disconnect the twitter",
	})
	return
}

func emojiHandler(c *gin.Context) {
	buf := make([]byte, 1024)
	c.Request.Body.Read(buf)
	trimmedData := bytes.TrimRight(buf, "\x00")
	var m map[string]interface{}
	err := json.Unmarshal(trimmedData, &m)
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"message": "invalid emoji message",
		})
		return
	}

	sender, senderOk := m["sender"].(string)
	signature, sigOk := m["signature"].(string)
	locationID, locationOk := m["message"].(map[string]interface{})["locationId"].(string)
	msgType, msgTypeOk := m["message"].(map[string]interface{})["type"].(string)
	bodyMap, bodyOk := m["message"].(map[string]interface{})["body"].(map[string]interface{})
	emoji, emojiOk := bodyMap["emoji"].(string)
	if !senderOk || !sigOk || !locationOk || !msgTypeOk || !bodyOk || !emojiOk {
		log.Println("invalid emoji message")
		c.JSON(302, gin.H{
			"success": false,
			"message": "invalid message",
		})
		return
	}

	verifyStr := fmt.Sprintf(`{"locationId":"%s","sender":"%s","type":"%s","body":{"emoji":"%s"}}`,
		locationID, sender, msgType, emoji)
	log.Println(verifyStr)
	realAddress, fullPubKey, err := generatePubKeyFromSign(signature, verifyStr)
	log.Println(hexutil.Encode(realAddress))
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"success": false,
			"message": err,
		})
		return
	}

	tmpSign, _ := hexutil.Decode(signature)
	userCheck := VerifySig(verifyStr, fullPubKey, tmpSign)
	log.Println("usercheck:", userCheck)
	if !userCheck {
		c.JSON(500, gin.H{
			"success": false,
			"message": err,
		})
		return
	}
	tmpM := make(map[string]string)
	tmpM["emoji"] = emoji
	tmpPlanetMessage := planetMessage{
		Id:          locationID,
		Sender:      sender,
		TimeCreated: time.Now().Unix(),
		PlanetId:    locationID,
		Body:        tmpM,
		EmojiFlag:   "EmojiFlag",
	}
	planetEmojiMap[locationID] = tmpPlanetMessage

	c.JSON(200, "success")
}

func emailHandler(c *gin.Context) {
	buf := make([]byte, 1024)
	c.Request.Body.Read(buf)
	trimmedData := bytes.TrimRight(buf, "\x00")
	var m map[string]interface{}
	err := json.Unmarshal(trimmedData, &m)
	if err != nil {
		c.JSON(302, gin.H{
			"success": false,
			"message": "invalid email info",
		})
		return
	}

	email := m["email"].(string)
	emailList = append(emailList, email)

	c.JSON(200, gin.H{
		"success": true,
		"message": "success to subscribe the email",
	})
	return
}

func messageHandler(c *gin.Context) {
	buf := make([]byte, 1024)
	c.Request.Body.Read(buf)
	trimmedData := bytes.TrimRight(buf, "\x00")
	var m map[string]interface{}
	err := json.Unmarshal(trimmedData, &m)
	if err != nil {
		c.JSON(302, gin.H{
			"success": false,
			"message": "invalid email info",
		})
		return
	}

	planets := m["planets"].([]interface{})
	res := make(map[string][]planetMessage)
	for i := range planets {
		planetId := planets[i].(string)
		pMsg, ok := planetEmojiMap[planetId]
		if ok {
			res[planetId] = make([]planetMessage, 1)
			res[planetId][0] = pMsg
		}
	}
	log.Println(res)
	c.JSON(200, res)
}
