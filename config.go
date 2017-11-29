package main

var mySigningKey = []byte("secret")
var jwtExpire = 3600

var db database = &DBBolt{}

const serverAddr = ":8000"
