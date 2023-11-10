package main

//test
import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/oklog/ulid/v2"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unicode/utf8"
)

type GetContentsData struct {
	Id         string `json:"id"`
	Curriculum string `json:"curriculum"`
	Category   string `json:"category"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Date       string `json:"datetime_column"`
}

type ContentsData struct {
	Curriculum string `json:"curriculum"`
	Category   string `json:"category"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Date       string `json:"datetime_column"`
}

// ① GoプログラムからMySQLへ接続
var db *sql.DB

func init() {

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
	w.Header().Set("Access-Control-Allow-Origin", "https://hackathon-front-one.vercel.app")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	//この行を入れたらエラーが消えた
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		curriculum := r.URL.Query().Get("curriculum")
		if curriculum == "home" || curriculum == "" {
			rows, err := db.Query("SELECT id, curriculum, category, title, body, datetime_column FROM contents")
			if err != nil {
				log.Printf("fail: db.Query, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			contentsdata := make([]GetContentsData, 0)
			for rows.Next() {
				var u GetContentsData
				if err := rows.Scan(&u.Id, &u.Curriculum, &u.Category, &u.Title, &u.Body, &u.Date); err != nil {
					log.Printf("fail: rows.Scan, %v\n", err)

					if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
						log.Printf("fail: rows.Close(), %v\n", err)
					}
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				contentsdata = append(contentsdata, u)
			}

			bytes, err := json.Marshal(contentsdata)
			if err != nil {
				log.Printf("fail: json.Marshal, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(bytes)
		} else {
			rows, err := db.Query("SELECT curriculum, category, title, body, datetime_column FROM contents WHERE curriculum = ?", curriculum)
			if err != nil {
				log.Printf("fail: db.Query, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			contentsdata := make([]ContentsData, 0)
			for rows.Next() {
				var u ContentsData
				if err := rows.Scan(&u.Curriculum, &u.Category, &u.Title, &u.Body, &u.Date); err != nil {
					log.Printf("fail: rows.Scan, %v\n", err)

					if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
						log.Printf("fail: rows.Close(), %v\n", err)
					}
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				contentsdata = append(contentsdata, u)
			}

			bytes, err := json.Marshal(contentsdata)
			if err != nil {
				log.Printf("fail: json.Marshal, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(bytes)
		}

	case http.MethodPost:
		// POSTメソッドの処理
		t := time.Now()
		entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
		id := ulid.MustNew(ulid.Timestamp(t), entropy)

		var requestData struct {
			Curriculum string `json:"curriculum"`
			Category   string `json:"category"`
			Title      string `json:"title"`
			Body       string `json:"body"`
			Date       string `json:"datetime_column"`
		}

		// HTTPリクエストボディからJSONデータを読み取る
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&requestData); err != nil {
			log.Printf("fail: json.Decode, %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if requestData.Category == "" {
			log.Println("fail: category is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if utf8.RuneCountInString(requestData.Category) > 50 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// データベースにINSERT
		_, err := db.Exec("INSERT INTO contents (id ,curriculum, category, title, body, datetime_column) VALUES (?,?,?,?,?,?)", id.String(), requestData.Curriculum, requestData.Category, requestData.Title, requestData.Body, requestData.Date)
		if err != nil {
			log.Printf("fail: db.Exec, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 成功した場合のレスポンス
		w.WriteHeader(http.StatusOK)
		response := map[string]string{"id": id.String()}
		bytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)

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
