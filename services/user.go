package services

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
	"github.com/wing8169/golang-htmx-chat-app/dto"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	DB *sql.DB
}

func (ts *UserService) GetUsers(username string) ([]*dto.UserDto, error) {
	users := []*dto.UserDto{}
	rows, err := ts.DB.Query("select id, username from user where username = ?", username)
	if err != nil {
		log.Println(err)
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var username string
		err = rows.Scan(&id, &username)
		if err != nil {
			log.Println(err)
			return users, err
		}
		users = append(users, &dto.UserDto{
			ID:       id,
			Username: username,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return users, err
	}
	return users, nil
}

func (ts *UserService) GetUser(username string) (*dto.UserDto, error) {
	stmt, err := ts.DB.Prepare("select id, password from user where username = ?")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()
	var id string
	var password string
	err = stmt.QueryRow(username).Scan(&id, &password)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &dto.UserDto{
		ID:       id,
		Username: username,
		Password: password,
	}, nil
}

func (ts *UserService) CreateUser(username string, password string) (*dto.UserDto, error) {
	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return nil, err
	}

	user := &dto.UserDto{
		ID:       uuid.New().String(),
		Username: username,
	}
	sqlStmt := `
	insert into user(id, username, password) values(?, ?, ?)
	`
	_, err = ts.DB.Exec(sqlStmt, user.ID, user.Username, string(hashedPassword))
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil, err
	}
	return user, nil
}

func (ts *UserService) UpdateUser(id string, username string) *dto.UserDto {
	user := &dto.UserDto{
		ID:       id,
		Username: username,
	}
	sqlStmt := `
	update user set username=? where id = ?
	`
	_, err := ts.DB.Exec(sqlStmt, user.Username, user.ID)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil
	}
	return user
}

func (ts *UserService) DeleteUser(id string) error {
	sqlStmt := `
	delete from user where id = ?
	`
	_, err := ts.DB.Exec(sqlStmt, id)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}

func (ts *UserService) LoginUser(username string, password string) (*dto.UserDto, error) {
	targetUser, err := ts.GetUser(username)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(targetUser.Password), []byte(password)); err != nil {
		return nil, err
	}

	return targetUser, nil
}
