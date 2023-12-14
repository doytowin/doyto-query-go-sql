package web

import (
	"github.com/doytowin/goquery/rdb"
	. "github.com/doytowin/goquery/test"
	log "github.com/sirupsen/logrus"
	"net/http/httptest"
	"testing"
)

func TestWeb(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	db := InitDB()
	createUserEntity := func() UserEntity { return UserEntity{} }
	userDataAccess := rdb.BuildRelationalDataAccess[UserEntity](createUserEntity)
	service := BuildService[rdb.Connection, UserEntity, UserQuery](
		"/user/", db, userDataAccess,
		createUserEntity,
		func() UserQuery { return UserQuery{} },
	)
	rs := &RestService[rdb.Connection, UserEntity, UserQuery]{Service: service}

	t.Run("Page /user/", func(t *testing.T) {
		writer := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/user/?PageNumber=1&PageSize=2", nil)

		rs.ServeHTTP(writer, request)

		actual := writer.Body.String()
		expect := `{"Data":{"List":[{"Id":1,"Score":85,"Memo":"Good"},{"Id":2,"Score":40,"Memo":"Bad"}],"Total":4},"Success":true,"Error":null}`
		if actual != expect {
			t.Errorf("\nExpected: %s\nBut got : %s", expect, actual)
		}
	})

	t.Run("Should return empty array instead of null when no data found.", func(t *testing.T) {
		writer := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/user/?PageNumber=10", nil)

		rs.ServeHTTP(writer, request)

		actual := writer.Body.String()
		expect := `{"Data":{"List":[],"Total":4},"Success":true,"Error":null}`
		if actual != expect {
			t.Errorf("\nExpected: %s\nBut got : %s", expect, actual)
		}
	})

	t.Run("Get /user/", func(t *testing.T) {
		writer := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/user/?ScoreLt=60&test=test", nil)

		rs.ServeHTTP(writer, request)

		actual := writer.Body.String()
		expect := `{"Data":{"List":[{"Id":2,"Score":40,"Memo":"Bad"},{"Id":3,"Score":55,"Memo":null}],"Total":2},"Success":true,"Error":null}`
		if actual != expect {
			t.Errorf("\nExpected: %s\nBut got : %s", expect, actual)
		}
	})

	t.Run("Get /user/ ", func(t *testing.T) {
		writer := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/user/?ScoreLt=60&MemoNull=true", nil)

		rs.ServeHTTP(writer, request)

		actual := writer.Body.String()
		expect := `{"Data":{"List":[{"Id":3,"Score":55,"Memo":null}],"Total":1},"Success":true,"Error":null}`
		if actual != expect {
			t.Errorf("\nExpected: %s\nBut got : %s", expect, actual)
		}
	})

	t.Run("Get /user/?MemoLike=%oo%", func(t *testing.T) {
		writer := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/user/?MemoLike=%25oo%25", nil)

		rs.ServeHTTP(writer, request)

		actual := writer.Body.String()
		expect := `{"Data":{"List":[{"Id":1,"Score":85,"Memo":"Good"}],"Total":1},"Success":true,"Error":null}`
		if actual != expect {
			t.Errorf("\nExpected: %s\nBut got : %s", expect, actual)
		}
	})

	t.Run("Get /user/?IdIn=1,4", func(t *testing.T) {
		writer := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/user/?IdIn=1,4", nil)

		rs.ServeHTTP(writer, request)

		actual := writer.Body.String()
		expect := `{"Data":{"List":[{"Id":1,"Score":85,"Memo":"Good"},{"Id":4,"Score":62,"Memo":"Well"}],"Total":2},"Success":true,"Error":null}`
		if actual != expect {
			t.Errorf("\nExpected: %s\nBut got : %s", expect, actual)
		}
	})

	t.Run("Get /user/1", func(t *testing.T) {
		writer := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/user/1", nil)

		rs.ServeHTTP(writer, request)

		actual := writer.Body.String()
		expect := `{"Data":{"Id":1,"Score":85,"Memo":"Good"},"Success":true,"Error":null}`
		if actual != expect {
			t.Errorf("\nExpected: %s\nBut got : %s", expect, actual)
		}
	})

	t.Run("Get /user/100", func(t *testing.T) {
		writer := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/user/100", nil)

		rs.ServeHTTP(writer, request)

		actual := writer.Body.String()
		expect := `{"Data":null,"Success":false,"Error":"record not found. id: 100"}`
		if actual != expect {
			t.Errorf("\nExpected: %s\nBut got : %s", expect, actual)
		}
	})
}
