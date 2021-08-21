package cookies

import (
	"database/sql"
	"fmt"
	"github.com/tlopo-go/cookie-exporter/decrypter"
	"github.com/tlopo-go/cookie-exporter/fileutils"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type fn func(string) error

func withTempDir(f fn) error {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	return f(tmp)
}

func Get() ([]*http.Cookie, error) {
	var cookies []*http.Cookie

	err := withTempDir(func(tmpDir string) error {

		cookies_db := `%s/Library/Application Support/Google/Chrome/Default/Cookies`
		cookies_db = fmt.Sprintf(cookies_db, os.Getenv("HOME"))

		ws_db := `%s/ws.db`
		ws_db = fmt.Sprintf(ws_db, tmpDir)

		fileutils.Copy(cookies_db, ws_db)

		db, err := sql.Open("sqlite3", ws_db)
		if err != nil {
			return err
		}

		rows, err2 := db.Query("SELECT host_key, path, expires_utc, name, value, encrypted_value FROM cookies ORDER by host_key")
		if err2 != nil {
			return err2
		}

		var host_key, path, expires_utc, name, value, encrypted_value string

		for rows.Next() {
			rows.Scan(&host_key, &path, &expires_utc, &name, &value, &encrypted_value)

			cookie := http.Cookie{
				Name:   name,
				Path:   path,
				Domain: host_key,
			}

			expires := getTimestamp(expires_utc)
			if expires > 0 {
				cookie.Expires = time.Unix(expires, 0)
			}

			getTimestamp(expires_utc)

			if len(encrypted_value) > 0 {
				decrypted, err3 := decrypter.Decrypt(encrypted_value)
				if err3 != nil {
					return err3
				}

				cookie.Value = strings.TrimSpace(decrypted)
			}

			cookies = append(cookies, &cookie)
		}

		return nil
	})
	return cookies, err
}

func GetNetscape() (netscape string, err error) {
	cookies, err := Get()

	if err != nil {
		return
	}

	var sb strings.Builder

	sb.WriteString("# Netscape HTTP Cookie File\n")
	for _, cookie := range cookies {
		expires := cookie.Expires.Unix()
		if expires < 0 {
			expires = 0
		}

		sb.WriteString(fmt.Sprintf(
			"%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			cookie.Domain, "TRUE", cookie.Path, "FALSE", expires, cookie.Name, cookie.Value,
		))
	}
	netscape = sb.String()
	return
}

func getTimestamp(timestr string) int64 {
	var ts int64

	ts, _ = strconv.ParseInt(timestr, 10, 64)

	if ts > 1000000 {
		ts = ts/1000000 - 11644473600
	}

	return ts
}
