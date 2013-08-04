#!/bin/bash
AUTH=`cat authfile`
xmllint --format <(curl --ntlm https://owa017.msoutlookonline.net/EWS/Exchange.asmx -u "$AUTH" --data @userAvailability.xml --header "content-type: text/xml; charset=utf-8")
