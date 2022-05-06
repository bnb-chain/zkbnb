/*
 * Copyright © 2021 Zecrey Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package util

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	curve "github.com/zecrey-labs/zecrey-crypto/ecc/ztwistededwards/tebn254"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"testing"
	"time"
)

func TestAccountNameHash(t *testing.T) {
	nameHash, err := AccountNameHash("sher.legend")
	if err != nil {
		panic(err)
	}
	fmt.Println(nameHash)
}

func TestPubKey(t *testing.T) {
	seed := "d892d866c5d0569e39e23c7bd46d63373d95197483e1a9af491e7098913a39ac"
	sk, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(common.Bytes2Hex(sk.PublicKey.Bytes()))
}

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

func TestRedis(t *testing.T) {
	r := redis.New("127.0.0.1:6379", WithRedis("node", "myredis"))
	_ = r.Set("key", "123")
	value, err := r.Get("key")
	if err != nil {
		panic(err)
	}
	fmt.Println(value)
	r.Del("key")
	redisLock := redis.NewRedisLock(r, "key")
	redisLock.SetExpire(2)
	isAcquired, err := redisLock.Acquire()
	if err != nil {
		panic(err)
	}
	if !isAcquired {
		panic("invalid key")
	}
	value, err = r.Get("key")
	fmt.Println(value)
	time.Sleep(time.Second * 3)
	isAcquired, err = redisLock.Acquire()
	if err != nil {
		panic(err)
	}
	if !isAcquired {
		panic("unable to acquire")
	}
	//_ = r.Set("key", "345")
	isReleased, err := redisLock.Release()
	if err != nil {
		panic(err)
	}
	if !isReleased {
		panic("unable to release")
	}
	value, err = r.Get("key")
	fmt.Println(value)

}

type Color struct {
	ColorType int64
}

func updateColors(colors map[string]*Color) {
	colors["0"] = &Color{
		2,
	}
}

func TestArray(t *testing.T) {
	var colors = make(map[string]*Color)
	updateColors(colors)
	fmt.Println(colors["0"])
}
