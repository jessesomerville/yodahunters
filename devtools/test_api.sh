#!/usr/bin/env -S bash -e

# Script for testing api endpoints after authz has been applied
# This just logs in and adds the cookie to a curl command to 
# make our lives a little easier.
# usage: ./devtools/test_api.sh username password [curl arguments ...]
token_regex="\"(eyJ.+)\""
username=$1
password=$2
shift 2
token_response=`curl -sS -X POST -d "{\"username\":\"$username\", \"password\":\"$password\"}" http://localhost:8080/api/login`
if [[ $token_response =~ $token_regex ]] 
then
    curl --cookie access_token=$BASH_REMATCH[1] "$@"
else
    echo "Failed to login!"
fi