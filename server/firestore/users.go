package firestore

import (
	"fmt"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

func (sd StockDocuments) SearchDocByProductID(productID string) (Stock, error) {
	for docID, stock := range sd {
		if stock.ProductID == productID {
			return sd[docID], nil
		}
	}
	return Stock{}, fmt.Errorf("couldn't find stock doc by: %s", productID)
}

func (sd StockDocuments) SearchDocIDByProductID(productID string) (string, error) {
	for docID, stock := range sd {
		if stock.ProductID == productID {
			return docID, nil
		}
	}
	return "", fmt.Errorf("couldn't find stock doc by: %s", productID)
}

func (client *Client) AddUser(profile *linebot.UserProfileResponse) (string, error) {
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return "", fmt.Errorf("couldn't create client in addUser: %v", err)
	}

	ref, _, err := c.Collection("users").Add(ctx, User{UserID: profile.UserID, DisplayName: profile.DisplayName, CreatedAt: time.Now()})
	if err != nil {
		return "", fmt.Errorf("couldn't find user document ref: %v", err)
	}

	return ref.ID, nil
}

func (client *Client) FetchUserByDocID(docID string) (User, error) {
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return User{}, fmt.Errorf("couldn't create client in fetchUser: %v", err)
	}

	var user User

	doc, err := c.Collection("users").Doc(docID).Get(ctx)
	if err != nil {
		return User{}, fmt.Errorf("couldn't find user document ref: %v", err)
	}

	if err := doc.DataTo(&user); err != nil {
		return User{}, err
	}
	// raw の user ID は LINE ID なので Doc ID に入れ替える。
	user.UserID = doc.Ref.ID
	return user, nil
}

func (client *Client) FetchUserByLINEProfile(profile *linebot.UserProfileResponse) (string, error) {
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return "", fmt.Errorf("couldn't create client in addUser: %v", err)
	}

	iter := c.Collection("users").Where("user_id", "==", profile.UserID).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			if doc == nil {
				// user がいなければ作成。
				userID, err := client.AddUser(profile)
				if err != nil {
					return "", err
				}
				return userID, nil
			}
			break
		}
		if err != nil {
			return "", err
		}
		return doc.Ref.ID, nil
	}

	return "", nil
}
