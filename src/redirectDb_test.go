package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const nameServiceDbAddr = "pub-redis-13248.us-east-mz.7.ec2.redislabs.com:13248"
const nameServiceDbPw = "=cLGW9&GQjCYu$2Z5+7K#6ED"
const redirectPrexif = "p6-qa"

func TestNS(t *testing.T) {
	ndb := newRedirectDb(nameServiceDbAddr, nameServiceDbPw, redirectPrexif)

	streams, err := ndb.streams()
	fmt.Println(streams)
	assert.Nil(t, err, "%s", err)

	assert.NotNil(t, streams, "a database struct didn't comeback")
}
