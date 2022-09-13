/*
 * Copyright © 2021 ZkBNB Protocol
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

package common

import (
	"bytes"
	"strings"
	"unicode"
)

func LowerCase(s string) string {
	return strings.ToLower(s)
}

func OmitSpace(s string) string {
	return strings.TrimSpace(s)
}

func OmitSpaceMiddle(s string) (rs string) {
	for _, v := range strings.FieldsFunc(s, unicode.IsSpace) {
		rs = rs + v
	}
	return rs
}

func CleanAccountName(name string) string {
	name = LowerCase(name)
	name = OmitSpace(name)
	name = OmitSpaceMiddle(name)
	return name
}

func SerializeAccountName(a []byte) string {
	return string(bytes.Trim(a[:], "\x00")) + ".legend"
}
