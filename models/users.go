package models

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"regexp"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"../hash"
	"../rand"
)

var (
	// ErrNotFound is returned when a resource can't be found inthe database
	ErrNotFound = errors.New("models: resource not found")

	// ErrInvalidID is returned when an invalid ID passed to a method like delete
	ErrIDInvalid = errors.New("models: ID Provided is invalid")

	// ErrPasswordIncorrect is returned when a user attempts to log in
	// with the wrong password
	ErrPasswordIncorrect = errors.New("models: incorrect password provided")

	// ErrEmailRequired is returned when an email address is not provided when creating a user
	ErrEmailRequired = errors.New("models: Email Address is Required")
	
	// ErrEmailInvalid is returned when an email address provided does not match our requirements
	ErrEmailInvalid = errors.New("models: Email Address is not valid.")

	// ErrEmailTaken is returned when the email is already in use
	ErrEmailTaken = errors.New("models: Email address is already taken")
	
	// ErrPasswordTooShort is returned when an update or create is 
	// attempted with a user password that is less than 8 characters
	ErrPasswordTooShort = errors.New("models: Password must be at least 8 characters long")

	// ErrPasswordRequired is returned when a create is attempted without
	// a user password
	ErrPasswordRequired = errors.New("models: Password is required")


	// match email addresses. not perfect but good enough
	emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@` + `[a-z0-9.\-]+\.[a-z]{2,16}$`)
)

const userPwPepper = "peter-picked-a-peck-of-pickled-peppers"
const hmacSecretKey = "secret-hmac-key"

// Represents the user model stored in our database
type User struct {
	gorm.Model 
	Name string
	Email string `gorm:"not null;unique_index"`
	Password string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

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

// UserService is a set of methods used to manipulate and work with the user model
type UserService interface {
	// Authenticate will verity the provided email address and password are correct.
	// If they are correct, the user corresponding to that email will be returned
	// otherwise you will recieve ErrNotFound,	ErrPasswordIncorrect or other error if something goes wrong
	Authenticate(email, password string) (*User, error)
	UserDB
}

func NewUserService(connectionInfo string) (UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
	  return nil, err
	}
	// this old line was in newUserGorm
	hmac := hash.NewHMAC(hmacSecretKey)
	uv := newUserValidator(ug, hmac)
	return &userService{
	  UserDB: uv,
	}, nil
  }

var _ UserService = &userService{}

type userService struct{
	UserDB
}

// Autheticate the user with an email and password
func (us *userService) Authenticate(email, password string) (*User, error){
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password + userPwPepper))
	if err != nil {	
		switch err{
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrPasswordIncorrect
		default:
			return nil, err
		}
	}
	return foundUser, nil
}

func runUserValFuncs(user *User, fns ...userValFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

var _ UserDB = &userValidator{}

func newUserValidator(udb UserDB, hmac hash.HMAC) *userValidator {
	return &userValidator{
		UserDB: udb,
		hmac: 	hmac, 
		emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@` + `[a-z0-9.\-]+\.[a-z]{2,16}$`),
	}
}

type userValidator struct {
	UserDB
	hmac hash.HMAC
	emailRegex *regexp.Regexp
}

// ByEmail will normalize the email address before calling ByEmail on the UserDB field
func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}
	if err := runUserValFuncs(&user, uv.normalizeEmail); err != nil{
		return nil, err
	}
	return uv.UserDB.ByEmail(user.Email)
}

// By remember will hash the remember token and call 
// ByRemember on the UserDB layer
func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User {
		Remember: token,
	}

	if err := runUserValFuncs(&user, uv.hmacRemember); err != nil {
		return nil, err
	}
	return uv.UserDB.ByRemember(user.RememberHash)
}

func (uv *userValidator) Create(user *User) error {
	err := runUserValFuncs(user, 
		uv.passwordRequired,
		uv.passwordMinLength,
		uv.bcryptPassword, 
		uv.setRememberIfUnset,
		uv.passwordHashRequired,
		uv.hmacRemember,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvail)
	if err != nil {
		return err
	}
	return uv.UserDB.Create(user)
}

// Update will update the provided user with all of the
// data in the user object
func (uv *userValidator) Update(user *User) error {
	err := runUserValFuncs(user,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.hmacRemember,
		uv.passwordHashRequired,
		uv.normalizeEmail,
		uv.emailFormat,
		uv.emailIsAvail)
	if err != nil {
		return err
	}
	return uv.UserDB.Update(user)
}

type userValFunc func(*User) error


func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

// Delete the user with the provided ID
func (uv *userValidator) Delete(id uint) error {
	// if you pass a 0 id, gorm will delete all users
	// we must check that the user exists 
	var user User
	user.ID = id

	err := runUserValFuncs(&user, uv.idGreaterThanZero )
   if err != nil {
	   return err
   }
	return uv.UserDB.Delete(id)
}

// bcryptPassword will has a users password with a predifined Pepper (userPwPepper)
// if the password field is not the empty string.
func (uv *userValidator) bcryptPassword(user *User) error {
	// only hash a password if it exists
	if user.Password == "" {
		return nil
	}
	// Add the password pepper
	pwBytes := []byte(user.Password + userPwPepper) 
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = "" // This isn't required, it's to prevent accidentally writing passwords to logs
	return nil
}

func (uv *userValidator) passwordMinLength(user *User) error {
	if user.Password == "" {
		return nil
	}
	if len(user.Password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

func (uv *userValidator) passwordRequired(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}

func (uv *userValidator) passwordHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}
	return nil
}

func (uv *userValidator) setRememberIfUnset(user *User) error {
	if user.Remember != ""{
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

func (uv *userValidator) idGreaterThanZero(user *User) error {
	// if you pass a 0 id, gorm will delete all users
	// we must check that the user exists 
	if user.ID <= 0 {
		return ErrIDInvalid
	}
	return nil
}

func (uv *userValidator) normalizeEmail(user *User) error {
	// clean up an email
	// make lowercase, trim whitespace
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func (uv *userValidator) emailFormat(user *User) error {
	if !uv.emailRegex.MatchString(user.Email){
		return ErrEmailInvalid
	}
	return nil
}

func (uv *userValidator) emailIsAvail(user *User) error {
	existing, err := uv.ByEmail(user.Email)
	if err == ErrNotFound {
		// EMail adderss is not taken
		return nil
	}
	if err != nil {
		return err
	}
	
	// We found a user with this email address
	// If the found user has the same ID as this user it is an update
	if user.ID != existing.ID {
		return ErrEmailTaken
	}
	return nil
}

func newUserGorm(connectionInfo string) (*userGorm, error){
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &userGorm{
		db: db, 
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

// ByRemember looks up a user with the given remember token
// and returns that user. This method expects the remember
// token to already be hashed.
func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
	err := first(ug.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
	  return nil, err
	}
	return &user, nil
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
	return ug.db.Create(user).Error
}


// Delete the user with the provided ID
func (ug *userGorm) Delete(id uint) error{
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(user.ID).Error
}

func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
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

