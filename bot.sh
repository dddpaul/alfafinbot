#!/bin/bash

GAS_URL="https://script.google.com/macros/s/${DEPLOYMENT_ID}/exec"

export TZ=Europe/Moscow
export GAS_URL="https://script.google.com/macros/s/AKfycbyI2cxgL6bY3aQGapJWBvGfE6mXwF8iZ2E3wlmMrM-N7wtkW-c7mckvX3_Uj5gyJT9ZZg/exec"

go run main.go \
-verbose \
-trace \
-telegram-admin dddpaul \
-telegram-token 5805269090:AAHUBxp5y-GjexLtpqSIWiCMyK29GMWrnLk \
-telegram-proxy-url "http://socks5master:fcZ%2333GK%213FC@Xjm@aruba.dddpaul.pw:11080" \
-gas-client-id alfafinbot \
-gas-client-secret eHa@!8X2A@d57%ab \
-gas-proxy-url "http://socks5master:fcZ%2333GK%213FC@Xjm@aruba.dddpaul.pw:11080"
