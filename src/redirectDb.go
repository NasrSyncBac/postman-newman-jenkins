package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Syncbak-Git/log"
	"github.com/garyburd/redigo/redis"
)

type Redirect struct {
	Streams   int       `json: "streams"`
	Max       int       `json: "max"`
	Host      string    `json: "host"`
	Timestamp time.Time `json:"timestamp"`
}

type redirectDb struct {
	server string
	pwd    string
	prefix string
}

func newRedirectDb(addr, pwd, prexif string) *redirectDb {
	return &redirectDb{server: addr, pwd: pwd, prefix: prexif}
}

func (r *Redirect) Since() int64 {
	return int64(time.Since(r.Timestamp) / time.Second)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

const redirectKey = "ns:redirect:"

func (db *redirectDb) streams() ([]*Redirect, error) {
	conn, err := db.connect()
	check(err)

	vals, err := redis.ByteSlices(conn.Do("HGETALL", fmt.Sprintf("%s%s", redirectKey, db.prefix)))
	check(err)

	var redirects []*Redirect

	for i, val := range vals {
		if i%2 == 1 {
			s := &Redirect{}
			err = json.Unmarshal(val, s)
			if err != nil {
				log.Error("Error unmarshalling redirect bytes %s error %s", string(val), err.Error())
				continue
			}
			if time.Since(s.Timestamp) < 2*time.Minute {
				redirects = append(redirects, s)
			}
		}
	}
	return redirects, nil
}

func (db *redirectDb) connect() (redis.Conn, error) {
	to := 2 * time.Second
	conn, err := redis.DialTimeout("tcp", db.server, to, to, to)
	check(err)

	if len(db.pwd) > 0 {
		if _, err := conn.Do("AUTH", db.pwd); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return conn, nil
}
