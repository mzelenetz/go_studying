package models

import (
	"fmt"
	"testing"
	"time"
)

func testingUserService() (*UserService, error){
	const (
		host = "localhost"
		port = 5432
		user = "postgres"
		password = "alexnoah"
		dbname = "databot_test"
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	"password=%s dbname=%s sslmode=disable",
	host, port, user, password, dbname)

	us, err := NewUserService(psqlInfo)
	if err != nil{
		return nil, err
	}
	us.db.LogMode(false)

	//Clear the users table between tests
	us.DestructiveReset()
	return us, nil
}

func TestCreateUser(t *testing.T) {
	us, err := testingUserService()
	if err != nil {
		t.Fatal(err)
	}
	
	user := User{
		Name: "Michael Scott",
		Email: "michael@dundermifflin.com",
	}	

	err = us.Create(&user)
	if err != nil {
		t.Fatal(err)
	}

	if user.ID == 0 {
		t.Errorf("Explected ID > 0. Received %d", user.ID)
	}

	if time.Since(user.CreatedAt) > time.Duration(5*time.Second){
		t.Errorf("Expected CreatedAt to be recent. Received %s", user.CreatedAt)
	}

	if time.Since(user.UpdatedAt) > time.Duration(5*time.Second){
		t.Errorf("Expected UpdatedAt to be recent. Received %s", user.UpdatedAt)
	}

	user.Email = "michael@michaelscottpaperco.com"
	if err := us.Update(&user); err != nil {
		t.Fatal(err)
	}

	// Get a new user object from the DB 
	userByID, err := us.ByID(user.ID)
	if err != nil{
		panic(err)
	}
	

	if userByID.Email != "michael@michaelscottpaperco.com"{
		t.Errorf("Expected Email to be michael@michaelscottpaperco.com. Received %s", userByID.Email)
	}

	// Get a new user object from the DB 
	userByEmail, err := us.ByEmail("michael@michaelscottpaperco.com")
	if err != nil{
		panic(err)
	}

	if userByEmail == nil {
		t.Errorf("Expected user to be found with email michael@michaelscottpaperco.com, none found.")
	}


}