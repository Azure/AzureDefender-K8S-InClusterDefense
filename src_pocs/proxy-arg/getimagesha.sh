#!/bin/sh

# registry := "upstream.azurecr.io"
# repo := "oss/kubernetes/ingress/nginx-ingress-controller"
# tag := "0.16.2"

usage() {
	echo "Invalid usage. Usage: "
	echo "\t$0 <registry> <repository> <tag>"
	exit 1
}
if [ $# -lt 3 ]; then
	usage
fi
registry="$1"
repository="$2"
tag="$3"
#echo $registry $repository $tag $CLIENT_ID

credentials=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
acr_access_token=$(curl -s -H "Content-Type: application/x-www-form-urlencoded" -H "Authorization: Basic $credentials" "https://$registry/oauth2/token?service=$registry&scope=repository:$repository:pull" | jq '.access_token' | sed -e 's/^"//' -e 's/"$//')

digest=$(curl -s -H "Authorization: Bearer $acr_access_token" https://$registry/acr/v1/$repository/_tags/$tag | jq .tag.digest | sed -e 's/^"//' -e 's/"$//')
if [ $? -ne 0 ] ; then
    echo $?
    exit 1
else
    echo $digest
	echo "Debug - digest was printed"
fi
exit 0
