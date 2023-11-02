package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type UserResForHTTPGet struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type ContentsData struct {
	Class string `json:"class"`
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
}

// ① GoプログラムからMySQLへ接続
var db *sql.DB

func init() {
	// ①-1
	// err := godotenv.Load(".env")
	// if err != nil {
	//	panic("Error loading .env file")
	//}

	// DB接続のための準備
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PWD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	connStr := fmt.Sprintf("%s:%s@%s/%s", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
	_db, err := sql.Open("mysql", connStr)

	// ①-2
	if err != nil {
		log.Fatalf("fail: sql.Open, %v\n", err)
	}
	// ①-3
	if err := _db.Ping(); err != nil {
		log.Fatalf("fail: _db.Ping, %v\n", err)
	}
	db = _db
}

// ② /userでリクエストされたらnameパラメーターと一致する名前を持つレコードをJSON形式で返す
func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// ②-1
		name := r.URL.Query().Get("name") // To be filled
		if name == "" {
			log.Println("fail: name is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		class := r.URL.Query().Get("name") // To be filled
		if class == "" {
			log.Println("fail: name is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// ②-2
		rows, err := db.Query("SELECT title, body, url FROM contents WHERE class = ?", class)
		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// ②-3
		contents := make([]ContentsData, 0)
		for rows.Next() {
			var u ContentsData
			if err := rows.Scan(&u.Class, &u.Title, &u.Body, &u.URL); err != nil {
				log.Printf("fail: rows.Scan, %v\n", err)

				if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
					log.Printf("fail: rows.Close(), %v\n", err)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			contents = append(contents, u)
		}

		// ②-4
		bytes, err := json.Marshal(contents)
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	case http.MethodPost:
		// POSTメソッドの処理
		//t := time.Now()
		//entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
		//id := ulid.MustNew(ulid.Timestamp(t), entropy)

		var requestData struct {
			Class string `json:"class"`
			Title string `json:"title"`
			Body  string `json:"body"`
			URL   string `json:"url"`
		}

		// HTTPリクエストボディからJSONデータを読み取る
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&requestData); err != nil {
			log.Printf("fail: json.Decode, %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if requestData.Class == "" {
			log.Println("fail: class is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if requestData.Title == "" {
			log.Println("fail: title is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(requestData.Class) > 50 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(requestData.Title) > 50 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// データベースにINSERT
		_, err := db.Exec("INSERT INTO user (class, title, body, url) VALUES (?,?,?,?)", requestData.Class, requestData.Title, requestData.Body, requestData.URL)
		if err != nil {
			log.Printf("fail: db.Exec, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 成功した場合のレスポンス
		w.WriteHeader(http.StatusOK)
		//response := map[string]string{"id": id.String()}
		//bytes, err := json.Marshal(response)
		//if err != nil {
		//	log.Printf("fail: json.Marshal, %v\n", err)
		//	w.WriteHeader(http.StatusInternalServerError)
		//	return
		//}
		//w.Header().Set("Content-Type", "application/json")
		//w.Write(bytes)

	default:
		log.Printf("fail: HTTP Method is %s\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func main() {
	// ② /userでリクエストされたらnameパラメーターと一致する名前を持つレコードをJSON形式で返す
	http.HandleFunc("/user", handler)

	// ③ Ctrl+CでHTTPサーバー停止時にDBをクローズする
	closeDBWithSysCall()

	// 8050番ポートでリクエストを待ち受ける
	log.Println("Listening...")
	if err := http.ListenAndServe(":8050", nil); err != nil {
		log.Fatal(err)
	}
}

// ③ Ctrl+CでHTTPサーバー停止時にDBをクローズする
func closeDBWithSysCall() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Printf("received syscall, %v", s)

		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		log.Printf("success: db.Close()")
		os.Exit(0)
	}()
}
