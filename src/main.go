// Copyright 2016 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
    // "encoding/json"
    "fmt"
    "github.com/google/go-github/github"
    "golang.org/x/oauth2"
)

var (
    user = "vmware"
    repo = "vic"
    myToken = "75413178dc833531cd42b157cec40ccb471de69d"
)

type TokenSource struct {
    AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
    token := &oauth2.Token{
        AccessToken: t.AccessToken,
    }
    return token, nil
}

func main() {
    tokenSource := &TokenSource{
        AccessToken: myToken,
    }

    oauthClient := oauth2.NewClient( oauth2.NoContext, tokenSource )
    client := github.NewClient(oauthClient)

    ilo := new(github.IssueListByRepoOptions)
    ilo.State = "all"
    ilo.PerPage = 1000
    ilo.Page = 1
    issues := []github.Issue{}
    iss, _, err := client.Issues.ListByRepo( user, repo, ilo )
    if err != nil {
        panic( err )
    }
    fmt.Println( "Got ", len(iss), " issues" )
    for len(iss) > 0 {
        ilo.Page++
        issues = append( issues, iss...)
        iss, _, err = client.Issues.ListByRepo( "vmware", "vic", ilo )
        if err != nil {
            panic( err )
        }
        fmt.Println( "Got ", len(iss), " issues" )
    }

    // fmt.Printf( "Issues:\n%s\n", string(e) ) */

    fmt.Println( len(issues), " issues reported:" )

    for _, iss := range issues {
        fmt.Print( "Issue # ", *iss.Number, "(", *iss.State, ") Assigned to: " )
        if iss.Assignee == nil {
            fmt.Print( "<no one>" )
        } else {
            fmt.Print( *iss.Assignee.Login )
        }

        for _, label := range iss.Labels {
            fmt.Print( " <", *label.Name, ">" )
        }
        fmt.Print( " Created: ", *iss.CreatedAt, " Closed: " )
        if ( iss.ClosedAt != nil ) {
            fmt.Println( *iss.ClosedAt )
        } else {
            fmt.Println( "<open>" )
        }
    }

    issues2 := []github.Issue{}
    for _, iss := range issues {
        for _, label := range iss.Labels {
            if *label.Name == "kind/bug" {
                issues2 = append( issues2, iss )
            }
        }
    }

    fmt.Println( len(issues2), " bugs reported" )
/*
    e, err := json.MarshalIndent( issues2, "", "  ")
    if err != nil {
        panic( err )
    }

    fmt.Printf( "Issue Detail: \n%s\n", string(e) )
*/
}
