package models

import (
	"errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	// ErrNotFound is returned when a resource can't be found inthe database
	ErrNotFound = errors.New("models: resource not found")

	// ErrInvalidID is returned when an invalid ID passed to a method like delete
	ErrInvalidID = errors.New("models: ID Provided is invalid")
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
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

// Look up a user object by the email
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

// Private method
// Need to pass a pointer to dst so that it can return results
// first will query using the gorm.DB object and it will retrieve
// the first item returned and place it into dst. 
// If no record is found it returns ErrNotFound
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error 
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	} 
	return err
}

// Create the provided user and will backfill data like
// id, created_at, updated_at fields
func (us *UserService) Create(user *User) error{
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = "" // This isn't required, it's to prevent accidentally writing passwords to logs
	return us.db.Create(user).Error
}

// Update will update the provided user with all of the
// data in the user object
func (us *UserService) Update(user *User) error {
	return us.db.Save(user).Error
}

// Delete the user with the provided ID
func (us *UserService) Delete(id uint) error{
	// if you pass a 0 id, gorm will delete all users
	// we must check that the user exists 
	if id == 0 {
		return ErrInvalidID
	}
	user := User{Model: gorm.Model{ID: id}}
	return us.db.Delete(user.ID).Error
}

// Closes the user service database connection
func (us *UserService) Close() error{
	return us.db.Close()
}

// Drops the user table and rebuilds it
func (us *UserService) DestructiveReset() error {
	if err := us.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	// Uses the automigrate we wrote in case we ever need to change the
	// AutoMigrate function
	return us.AutoMigrate()
}

// AutoMigrate will attempt to automatically migrate the users table
func (us *UserService) AutoMigrate() error {
	if err := us.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}

type User struct {
	gorm.Model 
	Name string
	Email string `gorm:"not null;unique_index"`
	Password string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
}