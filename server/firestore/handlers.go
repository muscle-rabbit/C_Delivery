package firestore

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func (client *Client) getUserHandler(g *gin.Context) {
	userID := g.Param("userID")
	userDocument, err := client.FetchUserByDocID(userID)
	if err != nil {
		g.Error(fmt.Errorf("couldn't fetchUser in getUserHandler: %v", err))
	}

	g.JSON(200, &userDocument)
}

func (client *Client) getOrdersHandler(g *gin.Context) {
	orderDocuments, err := client.FetchOrders()
	if err != nil {
		g.Error(fmt.Errorf("couldn't fetchOrders in orderListHandler: %v", err))
	}

	g.JSON(200, &orderDocuments)
	return
}

func (client *Client) getOrderHanlder(g *gin.Context) {
	orderID := g.Param("orderID")
	orderDocument, err := client.FetchUserOrder(orderID)
	if err != nil {
		g.Error(fmt.Errorf("coudln't fetchUserOder in getOrderHandler: %v", err))
	}

	g.JSON(200, &orderDocument)
}

func (client *Client) changeTradeStatusHandler(g *gin.Context) {
	action := g.Param("action")
	orderID := g.Param("orderID")
	switch action {
	case "/finish":
		if err := client.FinishTrade(orderID); err != nil {
			g.Error(fmt.Errorf("couldn't update order in finishTradeHandler: %v", err))
		}

	case "/unfinish":
		if err := client.UnfinishTrade(orderID); err != nil {
			g.Error(fmt.Errorf("couldn't update order in finishTradeHandler: %v", err))
		}
	default:
		g.Error(fmt.Errorf("coudln't parse action in finishTradeHandler"))
	}

	orderDoc, err := client.FetchUserOrder(orderID)
	if err != nil {
		g.Error(fmt.Errorf("couldn't fetch order in finishTradeHandler: %v", err))
	}
	g.JSON(200, &orderDoc)
}

func (client *Client) postChatHandler(g *gin.Context) {
	var message Message
	chatID := g.Param("chatID")
	err := g.BindJSON(&message)
	message.CreatedAt = time.Now()

	if err != nil {
		g.Error(fmt.Errorf("coudln't parse reader in postChatHandler: %v", err))
	}

	if err = client.PostChat(chatID, message); err != nil {
		g.Error(fmt.Errorf("coudln't post order chats in postChatHandler: %v", err))
	}
}

func (client *Client) getChatHandler(g *gin.Context) {
	chatID := g.Param("chatID")
	chatDoc, err := client.fetchChatDoc(chatID)
	if err != nil {
		g.Error(fmt.Errorf("coudln't fetch order chats in getChatHandler: %v", err))
	}

	g.JSON(200, &chatDoc)
}

func (m Menu) searchProductByName(key string) (ProductDocument, error) {
	for _, doc := range m {
		if doc.ProductItem.Name == key {
			return doc, nil
		}
	}
	// TODO: nil を返したい。
	return ProductDocument{}, fmt.Errorf("This product doesn't exist in Menu: %s", key)
}

func (m Menu) searcProductByID(id string) (ProductDocument, error) {
	for _, doc := range m {
		if doc.ID == id {
			return doc, nil
		}
	}
	return ProductDocument{}, fmt.Errorf("This product doesn't exist in Menu: %", id)
}

func (m Menu) calcPrice(products Products) int {
	var price int
	for id, product := range products {
		for _, v := range m {
			if v.ID == id {
				price += v.ProductItem.Price * product.Stock
			}
		}
	}
	return price
}
