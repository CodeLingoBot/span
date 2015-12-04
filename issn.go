//  Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//                    The Finc Authors, http://finc.info
//                    Martin Czygan, <martin.czygan@uni-leipzig.de>
//
// This file is part of some open source application.
//
// Some open source application is free software: you can redistribute
// it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation, either
// version 3 of the License, or (at your option) any later version.
//
// Some open source application is distributed in the hope that it will
// be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
// of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
//
// @license GPL-3.0+ <http://spdx.org/licenses/GPL-3.0+>
//
package span

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidISSN = errors.New("invalid ISSN")

var (
	replacer = strings.NewReplacer("-", "", " ", "")
	pattern  = regexp.MustCompile("^[0-9]{7}[0-9X]$")
)

type ISSN string

func (s ISSN) Validate() error {
	t := strings.TrimSpace(strings.ToUpper(replacer.Replace(string(s))))
	if len(t) != 8 {
		return ErrInvalidISSN
	}
	if !pattern.Match([]byte(t)) {
		return ErrInvalidISSN
	}
	return nil
}

func (s ISSN) String() string {
	t := strings.TrimSpace(strings.ToUpper(replacer.Replace(string(s))))
	if len(t) != 8 {
		return string(s)
	}
	return t[:4] + "-" + t[4:]
}