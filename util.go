package main

import (
    "log"
    "strings"
    "net/url"
    "errors"
    "encoding/base64"
)

func contains(leads []Lead, lead Lead) bool {
    for _, l := range leads {
        if (lead.Name == l.Name || lead.Url == l.Url) {
            return true
        }
    }

    return false
}


func formatUrl(str string) (string, error) {
    if (strings.Contains(str, "javascript")) {
        return str, errors.New("invalid url")
    }

    u, err := url.Parse(str)
    if (err != nil) {
        log.Fatal(err)
    }

    i, _ := base64.StdEncoding.DecodeString(strings.Split(u.Path, "/")[2])
    u2, _ := url.Parse(string(i))
    return u2.Path, nil
}