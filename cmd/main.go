package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"net/http"
	"os"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PutIp struct {
	Target string `json:"target"`
}

func main() {
	token := os.Getenv("LINODE_API_TOKEN")
	domainId := os.Getenv("LINODE_DOMAIN_ID")
	recordId := os.Getenv("LINODE_DOMAIN_RECORD_ID")

	url := fmt.Sprintf("https://api.linode.com/v4/domains/%s/records/%s", domainId, recordId)

	e := echo.New()

	cli := http.Client{}

	e.GET("/update", func(c echo.Context) error {
		b, _ := json.Marshal(&PutIp{Target: c.RealIP()})
		request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Response{1, err.Error()})
		}

		log.Printf("try to set ip=%s", c.RealIP())
		request.Header.Set("authorization", "Bearer "+token)

		response, err := cli.Do(request)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Response{2, err.Error()})
		}

		if response.StatusCode != http.StatusOK {
			r, _ := io.ReadAll(response.Body)
			return c.JSON(http.StatusInternalServerError, &Response{3, string(r)})
		}

		return c.JSON(http.StatusOK, &Response{0, "ok"})
	})

	e.Logger.Fatal(e.Start(":8888"))
}
