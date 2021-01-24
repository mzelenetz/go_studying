package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	// ErrNotFound is returned when a resource can't be found inthe database
	ErrNotFound = errors.New("models: resource not found")
)

func NewUserService(connectionInfo string) (*UserService, error){
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)

	return &UserService{
		db: db,
	}, nil
}

type UserService struct{
	db *gorm.DB

}

// ByID will look up a user by the id provided
// If the user is found: user, nil
// If the user is not found: nil, ErrNotFound
// If there is another error: nil, OtherError
func (us *UserService) ByID(id uint)(*User, error) {
	var user User
	err := us.db.Where("id = ?", id).First(&user).Error 
	switch err {
	case nil:
		return &user, nil
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Closes the user service database connection
func (us *UserService) Close() error{
	return us.db.Close()
}

// Drops the user table and rebuilds it
func (us *UserService) DestructiveReset(){
	us.db.DropTableIfExists(&User{})
	us.db.AutoMigrate(&User{})
}

type User struct {
	gorm.Model 
	Name string
	Email string `gorm:"not null;unique_index"`
}