// Copyright 2017 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	ignTypes "github.com/coreos/ignition/config/v2_0/types"
	"github.com/coreos/ignition/config/validate/report"
	"github.com/vincent-petithory/dataurl"
)

var (
	ErrUnknownGroup = errors.New("unknown update group")
)

type Update struct {
	Group  UpdateGroup  `yaml:"group"`
	Server UpdateServer `yaml:"server"`
}

type UpdateGroup string
type UpdateServer string

func (u Update) Validate() report.Report {
	switch strings.ToLower(string(u.Group)) {
	case "stable", "beta", "alpha":
		return report.Report{}
	default:
		if u.Server == "" {
			return report.ReportFromError(ErrUnknownGroup, report.EntryWarning)
		}
		return report.Report{}
	}
}

func (s UpdateServer) Validate() report.Report {
	_, err := url.Parse(string(s))
	if err != nil {
		return report.ReportFromError(err, report.EntryError)
	}
	return report.Report{}
}

func init() {
	register2_0(func(in Config, out ignTypes.Config, platform string) (ignTypes.Config, report.Report) {
		if in.Update != nil {
			out.Storage.Files = append(out.Storage.Files, ignTypes.File{
				Filesystem: "root",
				Path:       "/etc/coreos/update.conf",
				Mode:       0644,
				Contents: ignTypes.FileContents{
					Source: ignTypes.Url{
						Scheme: "data",
						Opaque: "," + updateConfContents(in.Update),
					},
				},
			})
		}
		return out, report.Report{}
	})
}

func updateConfContents(u *Update) string {
	var res string
	if u.Group != "" {
		res = fmt.Sprintf("GROUP=%s", strings.ToLower(string(u.Group)))
	}
	if u.Server != "" {
		res += fmt.Sprintf("\nSERVER=%s", u.Server)
	}
	return dataurl.EscapeString(res)
}