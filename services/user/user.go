package users

import (
	"context"
	"fmt"

	"main/common"
	"main/common/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID           string `json:"id" bson:"_id,omitempty"`
	Username     string `json:"username" bson:"username"`
	PasswordHash []byte `json:"-" bson:"passwordHash"`
	PasswordSalt []byte `json:"-" bson:"passwordSalt"`
	FullName     string `json:"fullName" bson:"fullName"`
	RoleName     string `json:"roleName" bson:"roleName"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"fullName"`
	RoleName string `json:"roleName"`
}

func CreateUser(db *db.DB, user User) error {
	_, err := GetUserByUsername(db, user.Username)
	if err == nil {
		return fmt.Errorf("username '%s' already exists", user.Username)
	}

	collection := db.Client().Database("PracticeDb").Collection("Users")
	_, err = collection.InsertOne(context.Background(), user)
	return err
}

func GetUserByID(db *db.DB, id string) (User, error) {
	var user User
	collection := db.Client().Database("PracticeDb").Collection("Users")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println("Error converting id to ObjectId:", err)
		return user, err
	}

	err = collection.FindOne(context.Background(), bson.M{"_id": objectId}).Decode(&user)
	if err != nil {
		fmt.Println("Error when GetUserByID(): ", err)
		return user, err
	}
	return user, nil
}

func GetUserByUsername(db *db.DB, username string) (User, error) {
	var user User
	collection := db.Client().Database("PracticeDb").Collection("Users")
	err := collection.FindOne(context.Background(), bson.M{"username": bson.M{"$regex": username, "$options": "i"}}).Decode(&user)
	return user, err
}

func UpdateUser(db *db.DB, id string, updateUser User) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"fullName": updateUser.FullName,
		},
	}

	collection := db.Client().Database("PracticeDb").Collection("Users")
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
	return err
}

func DeleteUser(db *db.DB, targetID string) error {
	targetObjectID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return err
	}

	_, err = GetUserByID(db, targetID)
	if err != nil {
		return err
	}

	collection := db.Client().Database("PracticeDb").Collection("Users")
	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": targetObjectID})
	return err
}

func GetUserList(db *db.DB, paging common.Paging) ([]User, error) {
	var users []User
	collection := db.Client().Database("PracticeDb").Collection("Users")

	findOptions := options.Find()
	findOptions.SetLimit(int64(paging.PageSize))
	findOptions.SetSkip(int64((paging.PageIndex - 1) * paging.PageSize))

	cursor, err := collection.Find(context.Background(), bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
