/* Copyright 2022 Zinc Labs Inc. and Contributors
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

package token

import (
	"fmt"
	"regexp"

	"github.com/blugelabs/bluge/analysis"

	"github.com/zinclabs/zincsearch/pkg/bluge/analysis/token"
	"github.com/zinclabs/zincsearch/pkg/errors"
	"github.com/zinclabs/zincsearch/pkg/zutils"
)

func NewRegexpTokenFilter(options interface{}) (analysis.TokenFilter, error) {
	pattern, err := zutils.GetStringFromMap(options, "pattern")
	if err != nil || pattern == "" {
		return nil, errors.New(errors.ErrorTypeParsingException, "[token_filter] regexp option [pattern] should be exists")
	}
	replacement, _ := zutils.GetStringFromMap(options, "replacement")
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, errors.New(errors.ErrorTypeParsingException, fmt.Sprintf("[token_filter] regexp option [pattern] compile error: %s", err.Error()))
	}

	return token.NewRegexpTokenFilter(r, []byte(replacement)), nil
}
