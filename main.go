package main

import (
	"net/http"
	"sort"
	"strings"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type (
	topTenWordsForm struct {
		Text string `json:"text" validate:"required"`
	}

	topTenWordsResponse struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
		Total   int         `json:"total"`
	}

	CustomValidator struct {
		validator *validator.Validate
	}
)

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code,
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func main() {
	r := echo.New()

	custValidator := &CustomValidator{validator: validator.New()}
	r.Validator = custValidator

	r.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status} latency=${latency_human} in:${bytes_in} out:${bytes_out}\n",
	}))

	r.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	r.Use(middleware.CORS())

	/** route /top-ten-words accept payload `text`
	  will return top ten most used words from `text`.
	  Response JSON example:
	  {
	    "success": true,
	    "data": [
	        {
	            "word": "aaa",
	            "total": 2
	        },
	        {
	            "word": "Go",
	            "total": 2
	        },
	        {
	            "word": "cc.",
	            "total": 1
	        },
	        {
	            "word": "The",
	            "total": 1
	        },
	        {
	            "word": "programming",
	            "total": 1
	        },
	        {
	            "word": "language",
	            "total": 1
	        },
	        {
	            "word": "is",
	            "total": 1
	        }
	    ],
	    "total": 7
	  } **/

	r.POST("/top-ten-words", func(ctx echo.Context) error {
		form := new(topTenWordsForm)
		if err := ctx.Bind(form); err != nil {
			return err
		}

		if err := ctx.Validate(form); err != nil {
			return err
		}

		mapWords := make(map[string]int)
		textLines := strings.Split(form.Text, "\n")
		for _, textLine := range textLines {
			words := strings.Split(textLine, " ")
			for _, word := range words {
				if len(word) < 1 || word[0] == '\n' {
					continue
				}

				if _, ok := mapWords[word]; ok {
					mapWords[word] += 1
				} else {
					mapWords[word] = 1
				}
			}
		}

		// Sort by value.
		type wordAndValue struct {
			Word  string `json:"word"`
			Value int    `json:"total"`
		}

		var wordsSort []wordAndValue
		for k, v := range mapWords {
			wordsSort = append(wordsSort, wordAndValue{k, v})
		}

		sort.Slice(wordsSort, func(i, j int) bool {
			return wordsSort[i].Value > wordsSort[j].Value
		})

		var topTenWords []wordAndValue
		for i := 0; i < 10; i++ {
			if i < len(wordsSort)-1 {
				topTenWords = append(topTenWords, wordsSort[i])
			}
		}

		resp := &topTenWordsResponse{
			Success: true,
			Data:    topTenWords,
			Total:   len(topTenWords),
		}
		return ctx.JSON(http.StatusOK, resp)
	})

	if err := r.Start(":9000"); err != nil {
		panic(err)
	}
}
