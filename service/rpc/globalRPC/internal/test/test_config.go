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
 */

package test

import (
	"flag"
	"fmt"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/config"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

func ConfigProvider() config.Config {
	logx.Info("ConfigProvider")

	// Parse .yaml file
	var c config.Config
	var configFile = flag.String("f", "../../etc/globalrpc.yaml", "the config file")
	fmt.Println(*configFile)
	flag.Parse()
	conf.MustLoad(*configFile, &c)
	return c
}
