package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Data struct {
	Title string
	Body  string
	Path string
	Action string
}

type Book struct {
	ISBN   string  `json:"isbn"`
	TITLE  string  `json:"title"`
	AUTHOR string  `json:"author"`
	PRICE  float32 `json:"price"`
}

type BookArray struct {
	Books []Book `json:"Book"`
}

type msgBook struct {
	ISBN    string `json:"isbn"`
	MESSAGE string `json:"message"`
}

var templates = template.Must(template.ParseGlob("templates/*"))

func main() {

	router := gin.Default()
	router.GET("/", IndexHandler)
	router.GET("/auth/", AuthHandler)
	router.POST("/auth/", getBooks)
	router.GET("/register/", RegisterHandler)
	router.POST("/register/", postAccount)
	router.GET("/about/", AboutHandler)
	router.GET("/books", getBooks)
	router.GET("/books/:isbn", getBookByID)
	router.POST("/books", postBooks)
	router.PUT("/books", putBooks)
	router.PATCH("/books/:isbn", patchBook)
	router.DELETE("/books/:isbn", deleteBook)
	log.Fatal(router.Run())

}

func DBConn() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgres://lcabral:T1QvJENu632ivVN56RuPjxXmQ2WlPOz4@dpg-ciatfc18g3nden787760-a/app_db_4wzw")
	//db, err := sql.Open("postgres", "postgres://lcabral:T1QvJENu632ivVN56RuPjxXmQ2WlPOz4@dpg-ciatfc18g3nden787760-a.ohio-postgres.render.com/app_db_4wzw")
	if err != nil {
		return db, err
	}

	if err = db.Ping(); err != nil {
		return db, err
	}

	return db, nil
}

func renderTemplate(ctx *gin.Context, tmpl string, page *Data) {
	err := templates.ExecuteTemplate(ctx.Writer, tmpl, page)
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func IndexHandler(ctx *gin.Context) {
	page := &Data{Title: "Home page", Body: "Welcome to our brand new home page.", Path: "/", Action: "Login"}
	renderTemplate(ctx, "index", page)
}

func AuthHandler(ctx *gin.Context) {
	page := &Data{Title: "Login page", Body: "Authentication", Path: ctx.FullPath(), Action: "Sign Up"}
	renderTemplate(ctx, "auth", page)
}

func RegisterHandler(ctx *gin.Context) {
	page := &Data{Title: "Register page", Body: "Registration account", Path: ctx.FullPath(), Action: "Register"}
	renderTemplate(ctx, "auth", page)
}

func AboutHandler(ctx *gin.Context) {
	page := &Data{Title: "About page", Body: "This is our brand new about page.", Path: ctx.FullPath(), Action: "Home"}
	renderTemplate(ctx, "index", page)
}

func postAccount(ctx *gin.Context) {
	db, err := DBConn()
	msg := ""

	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "auth.html", err.Error())
		return
	}

	//decodeJson := json.NewDecoder(ctx.Request.Body)
	firstname := ctx.PostForm("firstname")
	lastname := ctx.PostForm("lastname")
	email := ctx.PostForm("email")
	password := ctx.PostForm("password")

	//err = decodeJson.Decode(&req)

	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "auth.html", err.Error())
		return
	}
	
	_, err = db.Exec("INSERT INTO account VALUES($1, $2, $3, $4)", firstname, lastname, email, password)

	if err != nil {
		msg = err.Error()
		ctx.HTML(http.StatusInternalServerError, "auth.html", msg)
	} 

	msg = fmt.Sprintf("Account %s created successfully", email)
	ctx.HTML(http.StatusOK, "auth.html", msg)
}

func getBooks(ctx *gin.Context) {
	db, err := DBConn()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err})
	}
	rows, err := db.Query("SELECT * FROM books")

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err})
	}
	defer rows.Close()

	bks := make([]*Book, 0)

	for rows.Next() {
		bk := new(Book)
		err := rows.Scan(&bk.ISBN, &bk.TITLE, &bk.AUTHOR, &bk.PRICE)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err})
		}
		bks = append(bks, bk)
	}
	if err = rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err})
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"Books": bks})

}

func getBookByID(ctx *gin.Context) {
	isbn := ctx.Param("isbn")
	db, err := DBConn()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err})
	}

	row := db.QueryRow("SELECT * FROM books WHERE isbn = $1", isbn)

	bk := new(Book)

	err = row.Scan(&bk.ISBN, &bk.TITLE, &bk.AUTHOR, &bk.PRICE)

	if err == sql.ErrNoRows {
		ctx.JSON(http.StatusNotFound, gin.H{"Failed": "Book not found", "Erro": err.Error()})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, bk)

}

func putBooks(ctx *gin.Context) {
	db, err := DBConn()
	var req BookArray
	i := -1
	msg := ""

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err})
	}

	decodeJson := json.NewDecoder(ctx.Request.Body)

	err = decodeJson.Decode(&req)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err})
	}

	messages := make([]*msgBook, 0)

	for key, value := range req.Books {

		sql := "UPDATE books SET "

		if value.ISBN == "" {
			message := new(msgBook)
			message.ISBN = value.ISBN
			message.MESSAGE = fmt.Sprintf("The param ISBN does NOT is Null, title %s.", value.TITLE)
			messages = append(messages, message)
			continue
		} else {
			sql = sql + "isbn='" + value.ISBN + "'"
		}
		if value.TITLE != "" {
			sql = sql + ",title='" + value.TITLE + "'"
		}
		if value.AUTHOR != "" {
			sql = sql + ",author='" + value.AUTHOR + "'"
		}
		if value.PRICE != 0 {
			p := fmt.Sprintf("%v", value.PRICE)
			sql = sql + ",price=" + p
		}

		sql = sql + "  WHERE isbn = $1"

		_, err := db.Exec(sql, value.ISBN)
		if err != nil {
			msg = err.Error()
		} else {
			msg = fmt.Sprintf("Book %s altered successfully.", value.ISBN)
			i = key
		}
		message := new(msgBook)
		message.ISBN = value.ISBN
		message.MESSAGE = msg
		messages = append(messages, message)
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": messages})

	if i < 0 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"Failed": "Book NOT created, verify the request."})
	}
}

func patchBook(ctx *gin.Context) {
	db, err := DBConn()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err})
	}

	bk := new(Book)

	if err := ctx.ShouldBindJSON(&bk); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isbn := ctx.Param("isbn")
	title := bk.TITLE
	author := bk.AUTHOR
	price := bk.PRICE

	sql := "UPDATE books SET isbn='" + isbn + "'"

	if title != "" {
		sql = sql + ",title='" + title + "'"
	}
	if author != "" {
		sql = sql + ",author='" + author + "'"
	}
	if price != 0 {
		p := fmt.Sprintf("%v", price)
		sql = sql + ",price=" + p
	}

	sql = sql + "  WHERE isbn = $1"

	result, err := db.Exec(sql, isbn)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"Failed": "Book NOT created, verify the request."})
	} else {
		msg := fmt.Sprintf("Book %s altered successfully", isbn)
		ctx.IndentedJSON(http.StatusOK, gin.H{"success": msg})
	}

}

func deleteBook(ctx *gin.Context) {
	db, err := DBConn()
	isbn := ctx.Param("isbn")
	msg := ""

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err})
	}

	result, err := db.Exec("DELETE FROM books WHERE isbn = $1", isbn)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()

	if rowsAffected > 0 {
		msg = fmt.Sprintf("Book %s deleted successfully (%d row affected)", isbn, rowsAffected)
	} else {
		msg = fmt.Sprintf("Book %s NOT exists (%d row affected)", isbn, rowsAffected)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"success": msg})

}

func postBooks(ctx *gin.Context) {
	db, err := DBConn()
	var req BookArray
	i := -1
	msg := ""

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err.Error()})
	}

	decodeJson := json.NewDecoder(ctx.Request.Body)

	err = decodeJson.Decode(&req)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Failed": err.Error()})
		return
	}

	messages := make([]*msgBook, 0)

	for key, value := range req.Books {
		_, err := db.Exec("INSERT INTO books VALUES($1, $2, $3, $4)", value.ISBN, value.TITLE, value.AUTHOR, value.PRICE)
		if err != nil {
			msg = err.Error()
		} else {
			msg = fmt.Sprintf("Book %s created successfully.", value.ISBN)
			i = key
		}
		message := new(msgBook)
		message.ISBN = value.ISBN
		message.MESSAGE = msg
		messages = append(messages, message)
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": messages})

	if i < 0 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"Failed": "Book NOT created, verify the request."})
	}
}
