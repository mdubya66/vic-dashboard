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
    "bytes"
    "flag"
    "fmt"
    "io"
    "os"
    "sort"
    "github.com/google/go-github/github"
    "golang.org/x/oauth2"
    "net/html"
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

//
// Connect to github and iteratively pull all issues from the
// repository
//
func getAllIssues( user, repo *string, 
                   client *github.Client ) map[int]github.Issue {
    issues := make(map[int]github.Issue)

    ilo := new(github.IssueListByRepoOptions)
    ilo.State = "all"
    ilo.PerPage = 100
    ilo.Page = 1
    iss, _, err := client.Issues.ListByRepo( *user, *repo, ilo )
    if err != nil {
        panic( err )
    }
    for len(iss) > 0  {
        fmt.Printf( "Got %d issues\n", len(iss) )
        for _, i := range iss {
            issues[*i.Number] = i
        }
        ilo.Page++
        iss, _, err = client.Issues.ListByRepo( *user, *repo, ilo )
        if err != nil {
            panic( err )
        }
    }

    return issues
}

//
// Print all issues
//
func printAllIssues( issues map[int]github.Issue ) {
    keys := []int{}
    for k, _ := range issues {
        keys = append( keys, k )
    }
    sort.Ints(keys)

    for _, i := range keys {
        iss := issues[i]
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
    fmt.Println("")
}

func main() {

    //
    // Set up command-line parsing
    const (
        defaultUser = "vmware"
        defaultRepo = "vic"
        defaultTokenFile = ""
        defaultPort = 8888
    )

    myToken := os.Getenv( "VIC_DASHBOARD_TOKEN" )
    
    tokenFilePtr := flag.String( "tokenfile", 
                                 "", 
                                 "VIC Dashboard token file" )
    userPtr := flag.String( "user", defaultUser, "Github user" )
    repoPtr := flag.String( "repo", defaultRepo, "Github repo" )
    portPtr := flag.Int( "port", defaultPort, "TCP port to listen on" )

    flag.Parse()
    // End command-line parsing

    fmt.Println( myToken )
    fmt.Println( "TokenFile:  ", *tokenFilePtr )
    fmt.Println( "User:       ", *userPtr )
    fmt.Println( "Repo:       ", *repoPtr )
    fmt.Println( "Port:       ", *portPtr )

    if *tokenFilePtr != "" {
        buf := bytes.NewBuffer(nil)
        tf, err := os.Open( *tokenFilePtr )
        if err != nil {
            panic(err)
        }
        defer tf.Close()
        
        _, err = io.Copy( buf, tf )
        myToken = string(buf.Bytes())
    }
    fmt.Println( myToken )
    
    // Initialize the OAuth2 Token
    tokenSource := &TokenSource{
        AccessToken: myToken,
    }

    oauthClient := oauth2.NewClient( oauth2.NoContext, tokenSource )
    client := github.NewClient(oauthClient)
    issues := getAllIssues( userPtr, repoPtr, client )

    fmt.Println( len(issues), " issues reported:" )

    printAllIssues( issues )

}
