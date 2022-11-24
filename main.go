package main

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/kataras/iris/v12"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// varebles //
var DB *gorm.DB
var err error
var userName string
var sampleSecretKey = []byte("SecretYouShouldHide")

type User struct {
	ID         uint   `gorm:"primarykey;autoIncrement"`
	Username   string `gorm:"unique;not null""`
	Email      string
	Location   string
	Name       string
	Surname    string
	FatherName string
	BirthOfDay string
	Gender     string
	Number     uint
	Password   string
	Orders     []Order   `gorm:"foreignKey:Id`
	Products   []Product `gorm:"foreignKey:Id`
}
type Order struct {
	ID        uint `gorm:"primarykey;autoIncrement"`
	UserID    uint
	Productid int
	Price     float64
	Count     int
}
type Product struct {
	ID          uint `gorm:"primarykey;autoIncrement"`
	UserID      uint
	Title       string
	Description string
	Mainimage   string
	Images      string
	Rating      int
	Price       float64
	Salername   string
	Promotion   string
	Category    string
}
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func main() {
	app := iris.New()
	// root -> mysql-username
	// mysql -> mysql-password
	// (127.0.0.1:3306) -> host:port
	// golang-auth -> mysql-database-name
	dsn := "root:mysql@tcp(127.0.0.1:3306)/golang-auth"
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic("Mysql connect error!")
	}
	if !DB.Migrator().HasTable(&User{}) {
		DB.AutoMigrate(&User{})
	}
	if !DB.Migrator().HasTable(&Product{}) {
		DB.AutoMigrate(&Product{})
	}
	if !DB.Migrator().HasTable(&Order{}) {
		DB.AutoMigrate(&Order{})
	}
	//DB.AutoMigrate(&User{}, &Product{}, &Order{})
	app.Post("/api/register", register)
	app.Post("/api/login", login)
	app.Get("/user/{id:uint}", getUsersData)
	app.Post("/api/create/product", createProduct)
	app.Post("/api/create/order", createOrder)
	app.Delete("/api/product/{id}", deleteProduct)
	app.Delete("/api/order/{id}", deleteOrder)
	app.Listen(":8080")
}

// delete order
func deleteOrder(ctx iris.Context) {
	id := ctx.Params().Get("id")
	var order Order
	res := DB.Where("id = ?", id).Delete(&order)
	if res.Error != nil {
		ctx.JSON(iris.Map{
			"message": "Order no found!",
		})
		return
	}
	ctx.JSON(iris.Map{
		"message": "Order success deleted!",
	})
}

// delete product
func deleteProduct(ctx iris.Context) {
	id := ctx.Params().Get("id")
	var product Product
	res := DB.Where("id = ?", id).Delete(&product)
	if res.Error != nil {
		ctx.JSON(iris.Map{
			"message": "Order no found!",
		})
		return
	}
	ctx.JSON(iris.Map{
		"message": "Product success deleted!",
	})
}

// getUsersData handler
func getUsersData(ctx iris.Context) {
	var user User
	id := ctx.Params().Get("id")
	DB.Where(id).Find(&user)
	var product []Product
	var order []Order

	DB.Model(user).Association("Products").Find(&product)
	DB.Model(user).Association("Orders").Find(&order)
	ctx.JSON(iris.Map{
		"products": product,
		"username": user.Username,
		"email":    user.Email,
		"id":       user.ID,
		"orders":   order,
	})
}

// login handler
func login(ctx iris.Context) {
	var data struct {
		Username string `json: "username"`
		Password string `json: "password"`
	}
	var users struct {
		Id       uint
		Username string
		Email    string
		Password string
	}
	//var userData &users
	ctx.ReadJSON(&data)
	result := DB.Table("users").Where("username = ? and password = ?", data.Username, data.Password).Find(&users)
	if result.Error != nil {
		ctx.JSON(iris.Map{
			"message": "User not found!",
		})
		return
	}
	if users.Id == 0 {
		ctx.JSON(iris.Map{
			"message": "User data no correct!",
		})
		return
	}
	// create token //

	expirationTime := time.Now().Add(30 * time.Minute)
	claims := &Claims{
		Username: users.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		ctx.JSON(iris.Map{
			"message": "Created token error!",
		})
		return
	}
	ctx.SetCookieKV("token", tokenString)
	ctx.JSON(iris.Map{
		"message": "success",
		"token":   tokenString,
	})

}

// zakazat order handler
func createOrder(ctx iris.Context) {
	var data struct {
		Token     string  `json: "token"`
		Price     float64 `json: "price"`
		Productid int     `json: "productid"`
	}

	ctx.ReadJSON(&data)

	var tokenString string
	if ctx.GetCookie("token") == "" {
		tokenString = data.Token
	} else {
		tokenString = ctx.GetCookie("token")
	}
	claims := &Claims{}
	var sampleSecretKey = []byte("SecretYouShouldHide")
	tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return sampleSecretKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			ctx.JSON(iris.Map{
				"message": "Token is no invalid!",
			})
			return
		}
		ctx.JSON(iris.Map{
			"message": "No login!",
		})
		return
	}
	if !tkn.Valid {
		ctx.JSON(iris.Map{
			"message": "token parset error!",
		})
		return
	}
	userName = claims.Username
	var Datas User

	DB.Table("users").Where("username = ? ", userName).Find(&Datas)
	DB.Model(&Datas).Association("Orders").Append(&Order{
		Price:     data.Price,
		Productid: data.Productid,
	})

	ctx.JSON(iris.Map{
		"OrderId":  Datas.ID,
		"user":     Datas.Username,
		"username": userName,
	})
}

// createProduct handler
func createProduct(ctx iris.Context) {
	var data struct {
		Token       string  `json: "token"`
		Title       string  `json: "title"`
		Description string  `json: "description"`
		Mainimage   string  `json: "mainimage"`
		Images      string  `json: "image"`
		Rating      int     `json: "rating"`
		Price       float64 `json: "price"`
		Salername   string  `json: "salername"`
		Promotion   string  `json: "promotion"`
		Category    string  `json: "category"`
	}
	ctx.ReadJSON(&data)
	var tokenString string
	headerToken := ctx.GetCookie("token")
	if headerToken == "" {
		tokenString = data.Token
	} else {
		tokenString = headerToken
	}
	claims := &Claims{}
	var sampleSecretKey = []byte("SecretYouShouldHide")
	tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return sampleSecretKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			ctx.JSON(iris.Map{
				"message": "Token is no invalid!",
			})
			return
		}
		ctx.JSON(iris.Map{
			"message": "No login!",
		})
		return
	}
	if !tkn.Valid {
		ctx.JSON(iris.Map{
			"message": "token parset error!",
		})
		return
	}
	userName = claims.Username
	var Datas User

	DB.Table("users").Where("username = ? ", userName).Find(&Datas)
	DB.Model(&Datas).Association("Products").Append(&Product{
		Title:       data.Title,
		Description: data.Description,
		Price:       data.Price,
		Promotion:   data.Promotion,
		Mainimage:   data.Mainimage,
		Images:      data.Images,
		Salername:   Datas.Username,
		Rating:      data.Rating,
		Category:    data.Category,
	})

	ctx.JSON(iris.Map{
		"message":         "Product succses created!",
		"product_details": Datas,
	})
}

// regsiter handler
func register(ctx iris.Context) {
	var data struct {
		Username   string `json: "username"`
		Password   string `json: "password"`
		Email      string `json: "email"`
		Surname    string `json: "surname"`
		Name       string `json: "name"`
		FatherName string `json: "father_name"`
		Number     uint   `json: "number"`
		Gender     string `json: "gender"`
		Location   string `json: "location"`
		BirthOfDay string `json: "birth_of_day"`
	}
	ctx.ReadJSON(&data)
	result := DB.Create(&User{
		Username:   data.Username,
		Password:   data.Password,
		Email:      data.Email,
		Surname:    data.Surname,
		Location:   data.Location,
		FatherName: data.FatherName,
		Number:     data.Number,
		Name:       data.Name,
		Gender:     data.Gender,
		BirthOfDay: data.BirthOfDay,
	})
	if result.Error != nil {
		ctx.JSON(iris.Map{
			"message": "Failed data for create!",
		})
		return
	}
	expirationTime := time.Now().Add(30 * time.Minute)
	claims := &Claims{
		Username: data.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		ctx.JSON(iris.Map{
			"message": "Created token error!",
		})
		return
	}
	ctx.SetCookieKV("token", tokenString)
	ctx.JSON(iris.Map{
		"message": "succses",
		"token":   tokenString,
	})
}
