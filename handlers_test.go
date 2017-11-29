// handlers_test.go
package main

import (
	"bytes"
	"github.com/gavv/httpexpect"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

var AccessToken string
var testdb *DBBolt

func TestMain(t *testing.M) {
	log.Println("INIT")
	os.Remove("my.db")
	testdb = &DBBolt{}

	err := testdb.open()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new database
	err = testdb.init()
	if err != nil {
		log.Fatal(err)
	}
	retCode := t.Run()
	os.Exit(retCode)
}

func TestRegisterHandler1(t *testing.T) {
	var TEST_REQUEST = []byte(`{"Name":null,"Email":"test@example.com","MasterPasswordHash":"Q4zw5LmXHMJDJYBPfeFYtW8+dxbcCHTFmzE04OXS6Ic=","MasterPasswordHint":null,"Key":"0.yMeH5ypzRLcyJX69HAt6mQ==|H0mdMpoX1aguKIaCXOreL93JyCpo9ORiX8ZbK+taLXlGZfCb5TOs0eriKa7u1ocBp9gDHwYm5EUyobnbVfZ3uiP2suYWAXKmC4IO67b7ozc="}`)

	db = testdb

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/api/accounts/register", bytes.NewBuffer(TEST_REQUEST))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRegister)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := []byte{0x00}
	if !reflect.DeepEqual(rr.Body.Bytes(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestLoginHandler(t *testing.T) {
	var TEST_REQUEST = []byte(`grant_type=password&username=test%40example.com&password=Q4zw5LmXHMJDJYBPfeFYtW8%2BdxbcCHTFmzE04OXS6Ic%3D&scope=api+offline_access&client_id=mobile&DeviceType=Android&DeviceIdentifier=12345678-9abc-def0-1234-567890123456&DeviceName=Device&DevicePushToken=`)

	db = testdb
	handler := http.HandlerFunc(handleLogin)

	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
			Jar:       httpexpect.NewJar(),
		},
	})

	obj := e.POST("/identity/connect/token").
		WithHeader("Content-Type", "application/x-www-form-urlencoded").
		WithHeader("Accept", "application/json").
		WithBytes(TEST_REQUEST).
		Expect().
		Status(http.StatusOK).JSON().Object()

	AccessToken = obj.Path("$.access_token").String().Raw()
}

func TestSyncHandler(t *testing.T) {

	handler := jwtMiddleware(http.HandlerFunc(handleSync))

	db = testdb
	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
			Jar:       httpexpect.NewJar(),
		},
	})

	obj := e.GET("/api/sync").
		WithHeader("Authorization", "Bearer "+AccessToken).
		WithHeader("Accept", "application/json").
		Expect().
		Status(http.StatusOK).JSON().Object()

	if obj.Path("$.Profile.Email").Raw() != "test@example.com" {
		t.Errorf("email address doesn't match")
	}
}

/*
func TestNewCipherHandler(t *testing.T) {
	var TEST_REQUEST = []byte(`{"Type":1,"OrganizationId":null,"FolderId":null,"Favorite":false,"Name":"2.+pg5QYjJ4TSq6+atkQHGkg==|2p6Xj3zt/vysyPVe6FV+7Q==|wSEY6CfnJL80T4HCY3tXYX3CB3g2ejK4UWwO9FxxMyQ=","Notes":null,"Fields":null,"Login":{"Uri":"2.tzqIb8a3tQ5LHdT7Vtiyjw==|tlI+DKFaCzyGADujgOvilg==|sJtMIHxtf/Ai8VDDQ8p87D8zx2iTZwPYhlcWvg4rOF0=","Username":"2.sDX74yz5Kbst6y+xil9Vdw==|4H0JnjX9WKo9TBnf/JpcZQ==|sHIvZ3PH7CQN5nFAxjX5aXrzW9Nt/N3BKy8n+pU4Yas=","Password":"2.x+55fRdeu2b2fN2IlOBEdQ==|IO1LSR8kMhfmIFoPZy2A9w==|lFy55hvrR163YTNOxq+tngUQguORnzGpvo13JNE34hU=","Totp":null},"Card":null,"Identity":null,"SecureNote":null}`)

	db = testdb
	handler := jwtMiddleware(http.HandlerFunc(handleNewCipher))

	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
			Jar:       httpexpect.NewJar(),
		},
	})

	obj := e.POST("/api/ciphers").
		WithHeader("Authorization", "Bearer "+AccessToken).
		WithHeader("Content-Type", "application/json").
		WithHeader("Accept", "application/json").
		WithBytes(TEST_REQUEST).
		Expect().
		Status(http.StatusOK).JSON().Object()
	log.Println(obj)
}

func TestModifyCipherHandler(t *testing.T) {
	var TEST_REQUEST = []byte(`{"Type":1,"OrganizationId":null,"FolderId":null,"Favorite":false,"Name":"2.+pg5QYjJ4TSq6+atkQHGkg==|2p6Xj3zt/vysyPVe6FV+7Q==|wSEY6CfnJL80T4HCY3tXYX3CB3g2ejK4UWwO9FxxMyQ=","Notes":null,"Fields":null,"Login":{"Uri":"2.tzqIb8a3tQ5LHdT7Vtiyjw==|tlI+DKFaCzyGADujgOvilg==|sJtMIHxtf/Ai8VDDQ8p87D8zx2iTZwPYhlcWvg4rOF0=","Username":"2.sDX74yz5Kbst6y+xil9Vdw==|4H0JnjX9WKo9TBnf/JpcZQ==|sHIvZ3PH7CQN5nFAxjX5aXrzW9Nt/N3BKy8n+pU4Yas=","Password":"2.x+55fRdeu2b2fN2IlOBEdQ==|IO1LSR8kMhfmIFoPZy2A9w==|lFy55hvrR163YTNOxq+tngUQguORnzGpvo13JNE34hU=","Totp":null},"Card":null,"Identity":null,"SecureNote":null}`)

	db = testdb
	handler := jwtMiddleware(http.HandlerFunc(handleCipherUpdate))

	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
			Jar:       httpexpect.NewJar(),
		},
	})

	obj := e.PUT("/api/ciphers/1").
		WithHeader("Authorization", "Bearer "+AccessToken).
		WithHeader("Content-Type", "application/json").
		WithHeader("Accept", "application/json").
		WithBytes(TEST_REQUEST).
		Expect().
		Status(http.StatusOK).JSON().Object()
	log.Println(obj)
}
*/
