# vic-dashboard

This is a very early prototype of the VIC dashboard, built on Go.  It has been 
tested against Go-1.6.2, but should run on any current version of Go.  

Prerequisites:  In order to run this, you must have a OAuth2 token from Github.
This can either be set as an environment variable, VIC_DASHBOARD_TOKEN, or it can be
stored in a file and specified with the --tokenfile parameter.  The options available
are:

    --user=<user>      The github user to pull bugs from.  Default="vmware"
    --repo=<repo>      The github repo to pull bugs from.  Default="vic"
    --port=<port>      The (local) port to listen to.      Default=8888
    --tokenfile=<file> Filename containing OAuth2 token.   No default

The application will pull all issues from the vmware/vic repository and display them
on a web page at http://localhost:port

This is very much a work in progress.


