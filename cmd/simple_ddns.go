package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"gopkg.in/ini.v1"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PutIp struct {
	Target string `json:"target"`
}

type DDNS struct {
	Type           string `ini:"type"`
	APIToken       string `ini:"api_token"`
	DomainID       int    `ini:"domain_id"`
	DomainRecordID int    `ini:"domain_record_id"`
}

func main() {
	wd, _ := os.Getwd()
	configPath := path.Join(wd, "ddns.ini")
	log.Printf("configPath %s", configPath)

	conf := new(DDNS)
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		conf.Type = os.Getenv("DDNS_TYPE")
		conf.APIToken = os.Getenv("DDNS_API_TOKEN")
		conf.DomainID, _ = strconv.Atoi(os.Getenv("DDNS_DOMAIN_ID"))
		conf.DomainRecordID, _ = strconv.Atoi(os.Getenv("DDNS_DOMAIN_RECORD_ID"))
	} else {
		c := new(ini.File)
		if c, err = ini.Load(configPath); err != nil {
			panic(err)
		}
		if err = c.MapTo(&conf); err != nil {
			panic(err)
		}
	}

	url := fmt.Sprintf("https://api.linode.com/v4/domains/%d/records/%d", conf.DomainID, conf.DomainRecordID)

	p := fmt.Sprintf("%d:%d:%s", conf.DomainID, conf.DomainRecordID, conf.APIToken)
	s := sha256.Sum256([]byte(p))
	sid := hex.EncodeToString(s[:])

	log.Printf("id bind %s", sid)

	e := echo.New()
	e.GET("/update/:id", func(c echo.Context) error {
		id := c.Param("id")
		if id != sid {
			return c.JSON(http.StatusBadRequest, &Response{-1, "id not fit"})
		}

		b, _ := json.Marshal(&PutIp{Target: c.RealIP()})
		request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(b))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Response{1, err.Error()})
		}

		log.Printf("try to set ip=%s", c.RealIP())
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+conf.APIToken)

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Response{2, err.Error()})
		}

		r, _ := io.ReadAll(response.Body)
		if response.StatusCode != http.StatusOK {
			return c.JSON(http.StatusInternalServerError, &Response{3, string(r)})
		}

		log.Printf("dns provider response: %s", r)

		return c.JSON(http.StatusOK, &Response{0, "ok"})
	})

	e.Logger.Fatal(e.Start(":8888"))
}
