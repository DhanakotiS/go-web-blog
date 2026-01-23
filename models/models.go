package models

type User struct {
	Name     string `bson:"name"`
	Username string `bson:"username"`
	Password string `bson:"password"`
	Email    string `bson:"email"`
	UserID   uint32 `bson:"userid"`
}

type Blog struct {
	UserID  uint32 `bson:"userid"`
	BlogID  string `bson:"blogid"`
	Title   string `bson:"title"`
	Content string `bson:"content"`
}
