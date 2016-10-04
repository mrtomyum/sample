package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"golang.org/x/crypto/bcrypt"
)

const (
	DB_HOST = "tcp(nava.work:3306)"
	DB_NAME = "sample"
	DB_USER = "root"
	DB_PASS = "mypass"
)

type User struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name"`
	Password string `json:"password,omitempty"`
	Secret   []byte `json:"-" db:"secret"`
}

type People struct {
	ID int
	UserID int
	First string
	LastName string
	EmpCode string
	Sex string
}

type BcUser struct {
	ID int
	UserName string
	FirstName string
	EmpCode string
}

var db *sqlx.DB

func main() {
	var dsn = DB_USER + ":" + DB_PASS + "@" + DB_HOST + "/" + DB_NAME + "?parseTime=true"
	db = sqlx.MustConnect("mysql", dsn)

	r := gin.New()
	r.GET("/", Hello)
	r.GET("/users", SelectUsers)
	r.POST("/users", PostUser)
	r.POST("/users/login", UserLogin)

	r.Run(":8080")
}

func UserLogin(c *gin.Context) {
	loginUser := User{}
	if err := c.BindJSON(&loginUser); err != nil{
		c.String(400, err.Error())
	} else {
		existUser := User{}
		sql := `SELECT * FROM user WHERE name = ? LIMIT 1`
		db.Get(&existUser, sql, loginUser.Name)
		err := existUser.VerifyPass(loginUser.Password)
		if err != nil {
			c.String(200, err.Error())
			return
		} else {
			c.JSON(200, gin.H{"status":"SUCCESS"})
			return
		}
	}
}

func (u *User) VerifyPass(p string) error {
	err := bcrypt.CompareHashAndPassword(u.Secret, []byte(p))
	if err != nil {
		return err
	}
	return nil
}

func PostUser(c *gin.Context) {
	u := User{}
	if err := c.BindJSON(&u); err != nil{
		c.String(http.StatusBadRequest, err.Error())
	} else {
		log.Println(u)
		newUser, err := u.Insert(db)
		if err != nil {
			c.String(http.StatusNotImplemented, err.Error())
		}
		c.JSON(http.StatusOK, newUser)
	}
}

func (u *User) Insert(db *sqlx.DB) (*User, error) {
	sql := `
		INSERT INTO user(name, secret)
		VALUES(?,?)
	`
	u.SetPass()
	r, err := db.Exec(sql, u.Name, u.Secret)
	if err != nil {
		return nil, err
	}
	u.Password = ""
	id, _ := r.LastInsertId()
	var newUser User
	sql = `SELECT * FROM user WHERE id = ?`
	err = db.Get(&newUser,sql, id)
	if err != nil {
		return nil, err
	}
	return &newUser, nil
}

func (u *User) SetPass() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Secret = hash
	return nil
}

func SelectUsers(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Header("Access-Control-Allow-Origin", "*")

	var users []User
	sql := `SELECT name FROM user`
	err := db.Select(&users, sql)
	if err != nil {
		log.Println("Error Cannot Query", users, sql)
	}
	log.Println(users)
	c.JSON(http.StatusOK, users)
}

func Hello(c *gin.Context) {
	c.String(http.StatusOK, "Hello World")
}
