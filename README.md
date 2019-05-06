# AWS AppsSync Handler

## Resolver signatures

```text
func()
func() error
func(in) error
func() (out), error)
func(in) (out, error)
func(context.Context) error
func(context.Context, out) error
func(context.Context) (out, error)
func(context.Context, in) (out, error)
```

"in", "out" are types compatiable with the [encoding/json](https://golang.org/pkg/encoding/json).

## Example

### AppSync Request Mapping Template

```vtl
{
    "version": "2017-02-28",
    "operation": "Invoke",

    #set($args = $ctx.args.input)
    $utils.qr($args.put("userID", $ctx.identity.sub))

    "payload": {
        "resolve": "query.posts",
        "arguments": $utils.toJson($args)
    }
}
```

### Lambda function

```go
package main

import (
    "context"

    "github.com/bearchit/appsync-handler"
)


type postsInput struct {
    UserID    string `json:"userID"`
    Limit     uint64 `json:"limit"`
    NextToken string `json:"nextToken"`
}

type post struct {
    ID      string `json:"id"`
    Title   string `json:"title"`
    Content string `json:"content"`
}

func main() {
    h := appsync.NewHandler()

    h.AddResolver("query.post", func(ctx context.Context, input *postsInput) ([]*post, error) {
        // You can access `arguments` in the payload with struct `postInput`
        log.Println(input.UserID)
        log.Println(input.Limit)
        log.Println(input.NextToken)

        return []*post{
            {
                ID:      "1",
                Title:   "post #1",
                Content: "A content of post #1",
            },
            {
                ID:      "2",
                Title:   "post #2",
                Content: "A content of post #2",
            },
        }, nil
    })

    lambda.Start(h.Handle)
}
```
