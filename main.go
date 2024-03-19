package main

import (
    "context"
    "errors"
    "fmt"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson"
)

func main() {
    err := ConnectDB()
    if err != nil {
        fmt.Printf("MongoDB ga ulanishda xatolik: %v\n", err)
        return
    }

    err = addToCart("user123", "product123", 2)
    if err != nil {
        fmt.Printf("Savatchaga mahsulot qo'shishda xatolik: %v\n", err)
        return
    }

    err = removeFromCart("user123", "product123")
    if err != nil {
        fmt.Printf("Savatchadan mahsulotni olib tashlashda xatolik: %v\n", err)
        return
    }

    cartItems, err := viewCart("user123")
    if err != nil {
        fmt.Printf("Savatchadagi mahsulotlarni ko'rishda xatolik: %v\n", err)
        return
    }

    fmt.Println("Savatchadagi mahsulotlar:")
    for productID, quantity := range cartItems {
        fmt.Printf("Mahsulot ID: %s, Miqdori: %d\n", productID, quantity)
    }
}

const (
    mongoURI      = "mongodb://localhost:27017"
    databaseName  = "shopping_cart"
    usersCollName = "users"
    productsCollName = "products"
    cartsCollName = "carts"
)

var client *mongo.Client

type User struct {
    UserID   string `json:"user_id"`
    UserName string `json:"user_name"`
}

type Product struct {
    ProductID   string  `json:"product_id"`
    ProductName string  `json:"product_name"`
    Price       float64 `json:"price"`
}

type CartItem struct {
    UserID    string `json:"user_id"`
    ProductID string `json:"product_id"`
    Quantity  int    `json:"quantity"`
}

func ConnectDB() error {
    clientOptions := options.Client().ApplyURI(mongoURI)
    client, err := mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        return err
    }

    err = client.Ping(context.Background(), nil)
    if err != nil {
        return err
    }

    fmt.Println("Connected to MongoDB!")

    return nil
}

func addToCart(userID string, productID string, quantity int) error {
    if client == nil {
        return errors.New("MongoDB ga ulanilmagan")
    }

    cartItem := CartItem{
        UserID:    userID,
        ProductID: productID,
        Quantity:  quantity,
    }

    collection := client.Database(databaseName).Collection(cartsCollName)

    filter := bson.M{"user_id": userID, "product_id": productID}
    update := bson.M{"$set": cartItem}
    _, err := collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
    if err != nil {
        return err
    }

    return nil
}

func removeFromCart(userID string, productID string) error {
    if client == nil {
        return errors.New("MongoDB ga ulanilmagan")
    }

    collection := client.Database(databaseName).Collection(cartsCollName)

    filter := bson.M{"user_id": userID, "product_id": productID}
    _, err := collection.DeleteOne(context.Background(), filter)
    if err != nil {
        return err
    }

    return nil
}

func viewCart(userID string) (map[string]int, error) {
    if client == nil {
        return nil, errors.New("MongoDB ga ulanilmagan")
    }

    collection := client.Database(databaseName).Collection(cartsCollName)

    filter := bson.M{"user_id": userID}
    cursor, err := collection.Find(context.Background(), filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(context.Background())

    cartItems := make(map[string]int)
    for cursor.Next(context.Background()) {
        var cartItem CartItem
        if err := cursor.Decode(&cartItem); err != nil {
            return nil, err
        }
        cartItems[cartItem.ProductID] = cartItem.Quantity
    }
    if err := cursor.Err(); err != nil {
        return nil, err
    }

    return cartItems, nil
}
