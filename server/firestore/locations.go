package firestore

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

func (client *Client) fetchLocations() ([]Location, error) {
	var locations []Location
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't create client in getLocations: %v", err)
	}

	iter := c.Collection("locations").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var location Location
		if err := doc.DataTo(&location); err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}

	return locations, nil
}
