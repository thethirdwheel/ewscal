#!/bin/bash
AUTH=`cat ../data/authfile`
xmllint --format <(curl --ntlm https://owa017.msoutlookonline.net/EWS/Exchange.asmx -u "$AUTH" --data @$1 --header "content-type: text/xml; charset=utf-8")
