package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

type MyEvent struct {
	URL string `json:"url"`
}
type MyResponse struct {
	DATA  []map[string]interface{} `json:"data"`
	STATE bool                     `json:"state"`
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
// func Handler(ctx context.Context) (Response, error) {
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	myEvent := MyEvent{
		URL: "",
	}
	failResponseData := MyResponse{
		DATA:  nil,
		STATE: false,
	}
	failResponse, _ := json.Marshal(&failResponseData)

	err := json.Unmarshal([]byte(request.Body), &myEvent)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: string(failResponse), StatusCode: 200}, nil
	}

	fmt.Println("1 : " + myEvent.URL)
	res, err := http.Get("https://github.com" + myEvent.URL) //"laststance/create-react-app-typescript-todo-example-2021")
	if err != nil {
		//log.Fatal(err)
		return events.APIGatewayProxyResponse{Body: string(failResponse), StatusCode: 200}, nil
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		//log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		return events.APIGatewayProxyResponse{Body: string(failResponse), StatusCode: 200}, nil
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	result := []map[string]interface{}{}
	// Find the review items
	//.repo-content-pjax-container
	doc.Find(".js-details-container.Details div .Box-row").Each(func(i int, s *goquery.Selection) {
		row_types, ok := s.Find("div[role=gridcell] svg").Attr("aria-label")
		if !ok {
			fmt.Println("ERROR1", ok)
		}
		types := func(row_type string) string {
			if row_type == "File" {
				return "blob"
			}
			return "tree"
		}(row_types)

		name := s.Find("div[role=rowheader] span a").Text()
		link, ok := s.Find("div[role=rowheader] span a").Attr("href")
		if !ok {
			fmt.Println("ERROR2", ok)
		}
		if name != "" {
			result = append(
				result,
				map[string]interface{}{
					"type": types,
					"path": name,
					"url":  link,
				})
		}
	})

	successResponseData := MyResponse{
		DATA:  result,
		STATE: true,
	}
	successResponse, _ := json.Marshal(&successResponseData)

	// var buf bytes.Buffer
	body, err := json.Marshal(successResponseData)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: string(successResponse), StatusCode: 200}, nil
	}

	// json.HTMLEscape(&buf, body)
	// resp := Response{
	// 	StatusCode:      200,
	// 	IsBase64Encoded: false,
	// 	Body:            buf.String(),
	// 	Headers: map[string]string{
	// 		"Content-Type":           "application/json",
	// 		"X-MyCompany-Func-Reply": "hello-handler",
	// 	},
	// }

	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
