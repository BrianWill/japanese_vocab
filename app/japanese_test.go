// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	//	"net/http"
	//	"net/http/httptest"
	"database/sql"
	"fmt"
	"os"

	//"log"
	// "net/http"
	// "net/http/httptest"
	"testing"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// func TestIndexHandler(t *testing.T) {
// 	req, err := http.NewRequest("GET", "/", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(indexHandler)
// 	handler.ServeHTTP(rr, req)

// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf(
// 			"unexpected status: got (%v) want (%v)",
// 			status,
// 			http.StatusOK,
// 		)
// 	}

// 	expected := "Hello, World!"
// 	if rr.Body.String() != expected {
// 		t.Errorf(
// 			"unexpected body: got (%v) want (%v)",
// 			rr.Body.String(),
// 			"Hello, World!",
// 		)
// 	}
// }

// func TestIndexHandlerNotFound(t *testing.T) {
// 	req, err := http.NewRequest("GET", "/404", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(indexHandler)
// 	handler.ServeHTTP(rr, req)

// 	if status := rr.Code; status != http.StatusNotFound {
// 		t.Errorf(
// 			"unexpected status: got (%v) want (%v)",
// 			status,
// 			http.StatusNotFound,
// 		)
// 	}
// }

const USERHASH = "testuser"
const TEST_DB_PATH = "../users/" + USERHASH + ".db"

func setup(t *testing.T) {
	fmt.Println("testing: setup")
	var err error
	tok, err = tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}
	makeUserDB(USERHASH)
}

func teradown(t *testing.T) {
	fmt.Println("testing: teardown")
	e := os.Remove(TEST_DB_PATH)
	if e != nil {
		t.Fatal("could not teardown database")
	}
}

func TestAddAndGetStory(t *testing.T) {
	setup(t)
	defer teradown(t)

	sqldb, err := sql.Open("sqlite3", TEST_DB_PATH)
	if err != nil {
		t.Fatal("could not setup database")
	}
	defer sqldb.Close()

	story := Story{
		Title: "My Story",
		Link:  "http://youtube.com/asdf",
		Content: `
		0:15
		または10分します
		0:18
		ぜひ聴いてください今日のポッドキャスト
		0:21
		は小ストーリーです
		0:23
		日本語で物語を読みます
		0:26
		今日の物語は飴玉です
		0:29
		新美南吉さんが書いた物語です
		0:33
		この物語は jap pn 4レベルです
		0:37
		このエピソードのクイズがありますぜひ
		0:41
		クイズにチャレンジしてください
		0:45
		所ゆるゆりじゃないという
		0:47
		33カ所方法5
		0:50
		飴玉
		0:51
		春のとても暖かい日でした
		0:54
		お母さんと2人の子供がいました
		0:57
		3人は船に乗りました
		1:00
		船が出ようとすると男の声が聞こえました
		1:04
		多い
		1:05
		ちょっと待ってくれ`,
	}

	fmt.Println("testing: before add story")
	id, err := addStory(story, sqldb, false)
	if err != nil {
		t.Error("fail add story: ", err)
	}

	fmt.Println("testing: before get story")
	story, err = getStory(id, sqldb)
	if err != nil {
		t.Error("fail get story: ", err)
	}
}
