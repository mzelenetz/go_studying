package models

import (
	"errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"../hash"
	"../rand"
)

var (
	// ErrNotFound is returned when a resource can't be found inthe database
	ErrNotFound = errors.New("models: resource not found")

	// ErrInvalidID is returned when an invalid ID passed to a method like delete
	ErrInvalidID = errors.New("models: ID Provided is invalid")

	//
	ErrInvalidPassword = errors.New("models: incorrect password provided")
)

const userPwPepper = "peter-picked-a-peck-of-pickled-peppers"
const hmacSecretKey = "secret-hmac-key"

// This will be the database layer
// UserDB interacts with the User Database
// If the user is found: user, nil
// If the user is not found: nil, ErrNotFound
// If there is another error: nil, OtherError
//
// For single user queries, any error by ErrNotFound should probably
// result in a 500 error.
type UserDB interface {
	// Methods for querying for single users
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)
	
	// Methods for altering users
	Create(user *User) error 
	Update(user *User) error
	Delete(id uint) error 

	// Close a DB connection
	Close() error

	// Migration helpers
	// We don't have to put these in there. It's easier for development
	AutoMigrate() error
	DestructiveReset() error
}

func NewUserService(connectionInfo string) (*UserService, error){
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}

	return &UserService{
		UserDB: &UserValidator{
			UserDB: ug,
		}, 
	}, nil
}

type UserService struct{
	UserDB
}

type UserValidator struct {
	UserDB
}

func newUserGorm(connectionInfo string) (*userGorm, error){
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	hmac := hash.NewHMAC(hmacSecretKey)
	return &userGorm{
		db: db, 
		hmac: hmac,
	}, nil
}

type userGorm struct {
	db *gorm.DB
	hmac hash.HMAC
}


// ByID will look up a user by the id provided
// If the user is found: user, nil
// If the user is not found: nil, ErrNotFound
// If there is another error: nil, OtherError
func (ug *userGorm) ByID(id uint)(*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

// Look up a user object by the email
func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

// ByRemeber looks up a user with the given rememer token
// and reutrns that user. This method will handle hashing
// the token for us
// Errors are the same as ByEmail
func (ug *userGorm) ByRemember(token string) (*User, error) {
	var user User 
	rememberHash := ug.hmac.Hash(token)
	// query with the hashed token
	err := first(ug.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Autheticate the user with an email and password
func (us *UserService) Authenticate(email, password string) (*User, error){
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password + userPwPepper))
	if err != nil {	
		switch err{
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrInvalidPassword
		default:
			return nil, err
		}
	}
	return foundUser, nil
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
func (ug *userGorm) Create(user *User) error{
	// Add the password pepper
	pwBytes := []byte(user.Password + userPwPepper) 
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = "" // This isn't required, it's to prevent accidentally writing passwords to logs
	
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil  {
			return err
		}
		user.Remember = token
	}
	user.RememberHash = ug.hmac.Hash(user.Remember)	
	return ug.db.Create(user).Error
}

// Update will update the provided user with all of the
// data in the user object
func (ug *userGorm) Update(user *User) error {
	if user.Remember != ""{
		user.RememberHash = ug.hmac.Hash(user.Remember)
	}

	return ug.db.Save(user).Error
}

// Delete the user with the provided ID
func (ug *userGorm) Delete(id uint) error{
	// if you pass a 0 id, gorm will delete all users
	// we must check that the user exists 
	if id == 0 {
		return ErrInvalidID
	}
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(user.ID).Error
}

// Closes the user service database connection
func (ug *userGorm) Close() error{
	return ug.db.Close()
}

// Drops the user table and rebuilds it
func (ug *userGorm) DestructiveReset() error {
	if err := ug.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	// Uses the automigrate we wrote in case we ever need to change the
	// AutoMigrate function
	return ug.AutoMigrate()
}

// AutoMigrate will attempt to automatically migrate the users table
func (ug *userGorm) AutoMigrate() error {
	if err := ug.db.AutoMigrate(&User{}).Error; err != nil {
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
	Remember string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}