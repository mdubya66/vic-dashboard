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
    "encoding/base64"
    "github.com/google/go-github/github"
    "golang.org/x/oauth2"
    "golang.org/x/image/colornames"
    "net/http"

    "github.com/gonum/plot"
    "github.com/gonum/plot/plotter"
    "github.com/gonum/plot/vg"
)

type XYs plotter.XYs

type TokenSource struct {
    AccessToken string
}
type XY struct { X, Y float64 }
func (slice XYs) Len() int {
    return len(slice)
}

func (slice XYs) Less(i, j int) bool {
    return slice[i].X < slice[j].X
}
func (slice XYs) Swap(i, j int) {
    slice[i].X, slice[j].X = slice[j].X, slice[i].X
    slice[i].Y, slice[j].Y = slice[j].Y, slice[i].Y
}
func (slice XYs)  XY(i int)( x,y float64 ) {
    return slice[i].X, slice[i].Y
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
    token := &oauth2.Token{
        AccessToken: t.AccessToken,
    }
    return token, nil
}


var issues map[int]github.Issue

//
// Connect to github and iteratively pull all issues from the
// repository
//
func getAllIssues( user, repo *string, 
                   client *github.Client ) map[int]github.Issue {
    issues = make(map[int]github.Issue)

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

//
// Generate a graph of the bugs in terms of open, closed, and total
// Write these to a the http.ResponseWriter
//
func graphBugs(w http.ResponseWriter, r *http.Request ) {
    keys := []int{}
    for k, _ := range issues {
        keys = append( keys, k )
    }
    sort.Ints(keys)

    allBugs := XYs{}
    closedBugs := XYs{}
    openBugs := XYs{}
    
    point := XY{ 0.0, 1 }

    // Create data points for when each bug is open or closed
    for _, i := range keys {
        iss := issues[i]

        isBug := false
        for _, l := range iss.Labels {
            if *l.Name == "kind/bug" {
                isBug = true
            }
        }
        if isBug == false {
            continue
        }
/*
        fmt.Fprint( w, "Issue # ", *iss.Number, "(", *iss.State, ") Assigned to: " )
        if iss.Assignee == nil {
            fmt.Fprint( w, "<no one>" )
        } else {
            fmt.Fprint( w, *iss.Assignee.Login )
        }

        for _, label := range iss.Labels {
            fmt.Fprint( w, " <", *label.Name, ">" )
        }
        fmt.Fprint( w, " Created: ", *iss.CreatedAt, " Closed: " )
        if ( iss.ClosedAt != nil ) {
            fmt.Fprintln( w, *iss.ClosedAt )
        } else {
            fmt.Fprintln( w, "<open>" )
        } 
*/

        t := *iss.CreatedAt
        point.X = float64(t.Unix())
        allBugs = append( allBugs, point )
        if iss.ClosedAt != nil {
            t = *iss.ClosedAt
            point.X = float64( t.Unix() )
            closedBugs = append( closedBugs, point )
        }
    }

    // Sort bugs by time.
    // Technically allBugs should already be sorted by time,
    // but the closed date is only partially sorted
    sort.Sort( allBugs )
    sort.Sort( closedBugs )
    for i:=0; i<len(allBugs); i++ {
        allBugs[i].Y = float64(i+1)
    }
    for i:=0; i<len(closedBugs); i++ {
        closedBugs[i].Y = float64(i+1)
    }
    ab := 0  // count of all bugs
    cb := 0  // count of closed bugs
    ob := 0  // count of open bugs

    // Now let's create our X,Y points to count the open, closed, and all bugs
    for ab<len(allBugs) && cb<len(closedBugs) {
        if allBugs[ab].X < closedBugs[cb].X {
            ob++
            point = XY{ allBugs[ab].X, float64(ob) }
            ab++
        } else {
            ob--
            point = XY{ closedBugs[cb].X, float64(ob) }
            cb++
        }
        openBugs = append( openBugs, point )
    }
    if ab < len(allBugs) {
        for ab < len(allBugs) {
            ob++
            point = XY{ allBugs[ab].X, float64(ob) }
            openBugs = append( openBugs, point )
            ab++
        }
    } else {
        for cb < len(closedBugs) {
            ob--
            point = XY{ closedBugs[ab].X, float64(ob) }
            openBugs = append( openBugs, point )
            cb++
        }
    }

    // Now to plot the data
    p, err := plot.New()
    if err != nil {
        panic(err)
    }
    p.Title.Text = "vmware/vic Bug History"
    p.Y.Label.Text = "Number of Bugs"
    p.X.Tick.Marker = plot.UnixTimeTicks{Format: "2006-01-02"}
    p.Add(plotter.NewGrid())
    p.Legend.Top, p.Legend.Left = true, true

    line, err := plotter.NewLine(allBugs)
    if err != nil {
        panic(err)
    }
    line.Color = colornames.Blue
    p.Add( line )
    p.Legend.Add( "All Bugs", line )

    line, err = plotter.NewLine(closedBugs)
    if err != nil {
        panic(err)
    }

    line.Color = colornames.Green
    p.Add( line )
    p.Legend.Add( "Closed Bugs", line )

    line, err = plotter.NewLine(openBugs)
    if err != nil {
        panic(err)
    }

    line.Color = colornames.Red
    p.Add( line )
    p.Legend.Add( "Open Bugs", line )

    // Temporary spot to save to a PNG file
    err = p.Save(6*vg.Inch, 6*vg.Inch, "test.png")
    if err != nil {
        panic(err)
    }

    // Write a page to the http.ResponseWriter
    wr, err := p.WriterTo( 6 * vg.Inch, 6 * vg.Inch, "png" )
    if err != nil {
        panic(err)
    }
    
    // Encode it as base64
    var b bytes.Buffer
    b64w := base64.NewEncoder( base64.StdEncoding, &b )
    wr.WriteTo( w )
    b64w.Close()
    str := string( b.Bytes() )

    fmt.Fprint( w,
        `<!DOCTYPE HTML>
         <html>
         <head>
         <title>VIC Bug History</title>
         </head>

         <body>
         <img alt="VIC Bug History Image" src="data:image/png;charset=utf8;base64, ` )
    fmt.Fprint( w, str )
    fmt.Fprintln( w, `" />` )
    fmt.Fprintln( w, `<body></html` )
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

    http.HandleFunc( "/", graphBugs )
    serverPort := fmt.Sprintf( ":%d", *portPtr )
    fmt.Println( "Listening on port ", serverPort )
    http.ListenAndServe( serverPort, nil )
}
