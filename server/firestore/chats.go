package firestore

import (
	"fmt"

	"cloud.google.com/go/firestore"
	"golang.org/x/net/context"
)

func (client *Client) fetchChatDoc(chatID string) (ChatDocument, error) {
	var chatDoc ChatDocument
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return ChatDocument{}, fmt.Errorf("couldn't create client in fetchOrderChats: %v", err)
	}
	doc, err := c.Collection("chats").Doc(chatID).Get(ctx)
	if err != nil {
		return ChatDocument{}, err
	}
	if err = doc.DataTo(&chatDoc); err != nil {
		return ChatDocument{}, err
	}
	return chatDoc, nil
}

func (client *Client) PostChat(chatID string, message Message) error {
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in post: %v", err)
	}

	chat := c.Collection("chats").Doc(chatID)
	_, err = chat.Update(ctx, []firestore.Update{
		{Path: "messages", Value: firestore.ArrayUnion(message)},
	})
	if err != nil {
		return err
	}
	return err
}

func (client *Client) createChatRoom(orderID string, userID string) (string, error) {
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return "", fmt.Errorf("couldn't create client in post: %v", err)
	}

	ref, _, err := c.Collection("chats").Add(ctx, map[string]interface{}{
		"order_id": orderID,
	})
	if err != nil {
		return "", err
	}
	return ref.ID, err
}
